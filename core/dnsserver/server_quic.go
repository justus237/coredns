package dnsserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
)

var log_ = log.NewWithPlugin("DoQ")

// maxQuicIdleTimeout - maximum QUIC idle timeout.
// Default value in quic-go is 30, but our internal tests show that
// a higher value works better for clients written with ngtcp2
const maxQuicIdleTimeout = 5 * time.Minute

const minDNSPacketSize = 12 + 5

// Implemented according to https://tools.ietf.org/html/draft-huitema-dprive-dnsoquic-00
// ServerQUIC represents an instance of a DNS-over-QUIC server.
type ServerQUIC struct {
	*Server
	tlsConfig  *tls.Config
	listen     quic.Listener
	listenAddr net.Addr

	bytesPool *sync.Pool
}

// NewServerQUIC returns a new CoreDNS QUIC server and compiles all plugin in to it.
func NewServerQUIC(addr string, group []*Config) (*ServerQUIC, error) {
	s, err := NewServer(addr, group)
	if err != nil {
		return nil, err
	}
	// The *tls* plugin must make sure that multiple conflicting
	// TLS configuration return an error: it can only be specified once.
	var tlsConfig *tls.Config
	for _, conf := range s.zones {
		// Should we error if some configs *don't* have TLS?
		tlsConfig = conf.TLSConfigQUIC
	}

	bytesPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, dns.MaxMsgSize)
		},
	}

	log_.Info("NewServerQUIC()")
	return &ServerQUIC{Server: s, tlsConfig: tlsConfig, bytesPool: &bytesPool}, nil
}

// Compile-time check to ensure Server implements the caddy.GracefulServer interface
var _ caddy.GracefulServer = &Server{}

// Serve implements caddy.TCPServer interface.
func (s *ServerQUIC) Serve(_ net.Listener) error {
	return nil
}

// ServePacket implements caddy.UDPServer interface.
func (s *ServerQUIC) ServePacket(p net.PacketConn) error {
	s.m.Lock()
	log_.Info("ServePacket()")

	if s.tlsConfig == nil {
		return errors.New("cannot run a QUIC server without TLS config")
	}
	var customAcceptToken = func(clientAddr net.Addr, token *quic.Token) bool {
		/*log_.Infof("token acceptor called for: %s\n", clientAddr.String())
		if token == nil {
			log_.Infof("no token, rejecting and asking for retry\n")
			return false
		}
		log_.Infof("token with remote addr: %s\n", token.RemoteAddr)*/
		return true
	}
	l, err := quic.Listen(p, s.tlsConfig, &quic.Config{MaxIdleTimeout: maxQuicIdleTimeout, AcceptToken: customAcceptToken})
	//l, err := quic.Listen(p, s.tlsConfig, &quic.Config{MaxIdleTimeout: maxQuicIdleTimeout, StatelessResetKey: nil})
	if err != nil {
		return err
	}
	s.listen = l
	s.listenAddr = l.Addr()
	s.m.Unlock()

	for {
		session, err := s.listen.Accept(context.Background())
		if err != nil {
			return err
		}

		go s.handleQUICSession(session)
	}
}

// Listen implements caddy.TCPServer interface.
func (s *ServerQUIC) Listen() (net.Listener, error) { return nil, nil }

// ListenPacket implements caddy.UDPServer interface.
func (s *ServerQUIC) ListenPacket() (net.PacketConn, error) {
	p, err := reuseport.ListenPacket("udp", s.Addr[len(transport.QUIC+"://"):])
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Stop stops the server. It blocks until the server is totally stopped.
func (s *ServerQUIC) Stop() error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.listen.Close()
}

// OnStartupComplete lists the sites served by this server
// and any relevant information, assuming Quiet is false.
func (s *ServerQUIC) OnStartupComplete() {
	if Quiet {
		return
	}

	out := startUpZones(transport.QUIC+"://", s.Addr, s.zones)
	if out != "" {
		fmt.Print(out)
	}
}

func (s *ServerQUIC) handleQUICSession(session quic.Session) {
	for {
		// The stub to resolver DNS traffic follows a simple pattern in which
		// the client sends a query, and the server provides a response.  This
		// design specifies that for each subsequent query on a QUIC connection
		// the client MUST select the next available client-initiated
		// bidirectional stream
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			_ = session.CloseWithError(0, "")
			return
		}
		go func() {
			s.handleQUICStream(stream, session)
			_ = stream.Close()
		}()
	}
}

// handleQUICStream reads DNS queries from the stream, processes them,
// and writes back the responses
func (s *ServerQUIC) handleQUICStream(stream quic.Stream, session quic.Session) {
	var b []byte
	b = s.bytesPool.Get().([]byte)
	defer s.bytesPool.Put(b)

	// The client MUST send the DNS query over the selected stream, and MUST
	// indicate through the STREAM FIN mechanism that no further data will
	// be sent on that stream.
	// FIN is indicated via error so we should simply ignore it and
	// check the size instead.
	n, _ := stream.Read(b)
	if n < minDNSPacketSize {
		// Invalid DNS query, this stream should be ignored
		return
	}

	msg := new(dns.Msg)
	err := msg.Unpack(b)
	if err != nil {
		// Invalid content
		return
	}

	// If any message sent on a DoQ connection contains an edns-tcp-keepalive EDNS(0) Option,
	// this is a fatal error and the recipient of the defective message MUST forcibly abort
	// the connection immediately.
	// https://datatracker.ietf.org/doc/html/draft-ietf-dprive-dnsoquic-02#section-6.6.2
	if opt := msg.IsEdns0(); opt != nil {
		for _, option := range opt.Option {
			// Check for EDNS TCP keepalive option
			if option.Option() == dns.EDNS0TCPKEEPALIVE {
				// Already closing the connection so we don't care about the error
				_ = session.CloseWithError(0, "")
			}
		}
	}

	// Consider renaming DoHWriter or creating a new struct for QUIC
	dw := &DoHWriter{laddr: s.listenAddr, raddr: session.RemoteAddr()}

	// We just call the normal chain handler - all error handling is done there.
	// We should expect a packet to be returned that we can send to the client.
	ctx := context.WithValue(context.Background(), Key{}, s.Server)
	s.ServeDNS(ctx, dw, msg)

	if dw.Msg == nil {
		_ = stream.Close()
		return
	}

	// Write the response
	buf, _ := dw.Msg.Pack()
	_, _ = stream.Write(buf)
}
