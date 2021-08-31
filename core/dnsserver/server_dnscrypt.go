package dnsserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/ameshkov/dnscrypt/v2"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/miekg/dns"
)

// It seems that UDP size 1252 is the default one used by dnscrypt-proxy so
// it's reasonable to have it as the default read buffer size:
// https://github.com/AdguardTeam/AdGuardDNS/issues/188
const defaultUDPSize = 1252

// ServerDNSCrypt implement a server for processing requests using DNSCrypt protocol.
type ServerDNSCrypt struct {
	*Server
	dnsCryptServer *dnscrypt.Server
}

// dnsCryptHandler implement dnscrypt.Handler
type dnsCryptHandler struct {
	server *Server
}

func (h *dnsCryptHandler) ServeDNS(rw dnscrypt.ResponseWriter, req *dns.Msg) error {

	// Consider renaming DoHWriter or creating a new struct for DNSCrypt.
	nonw := &DoHWriter{nonwriter.Writer{}, rw.LocalAddr(), rw.RemoteAddr()}

	// We just call the normal chain handler - all error handling is done there.
	// We should expect a packet to be returned that we can send to the client.
	ctx := context.WithValue(context.Background(), Key{}, h.server)
	h.server.ServeDNS(ctx, nonw, req)

	msg := nonw.Msg
	if msg == nil {
		return fmt.Errorf("dw.Msg is empty")
	}
	return rw.WriteMsg(msg)
}

var (
	ErrEmptyConfig = errors.New("DNSCrypt config is empty. " +
		"You must specify the path to the configuration " +
		"using the 'dnscrypt' corfile directive")
)

func NewServerDNSCrypt(addr string, group []*Config) (*ServerDNSCrypt, error) {
	s, err := NewServer(addr, group)
	if err != nil {
		return nil, err
	}

	// resolverConfig taken from dnsserver.Config.DNSCryptConfig.
	// Is populated by the dnscrypt plugin from yaml file or Corefile.
	var resolverConfig *dnscrypt.ResolverConfig
	for _, conf := range s.zones {
		resolverConfig = conf.DNSCryptConfig
		if resolverConfig != nil {
			break
		}
	}
	if resolverConfig == nil {
		return nil, ErrEmptyConfig
	}

	cert, err := resolverConfig.CreateCert()
	if err != nil {
		return nil, err
	}

	return &ServerDNSCrypt{
		Server: s,
		dnsCryptServer: &dnscrypt.Server{
			ProviderName: resolverConfig.ProviderName,
			ResolverCert: cert,
			Handler: &dnsCryptHandler{
				server: s,
			},
			UDPSize: defaultUDPSize,
		},
	}, nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information, assuming Quiet is false.
func (s *ServerDNSCrypt) OnStartupComplete() {
	if Quiet {
		return
	}

	out := startUpZones(transport.DNSCrypt+"://", s.Addr, s.zones)
	if out != "" {
		fmt.Print(out)
	}
}

func (s ServerDNSCrypt) Listen() (net.Listener, error) {
	addr := s.Addr[len(transport.DNSCrypt+"://"):]
	l, err := reuseport.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (s ServerDNSCrypt) Serve(l net.Listener) error {
	for {
		err := s.dnsCryptServer.ServeTCP(l)
		if err != nil {
			return err
		}
	}
}

func (s ServerDNSCrypt) ListenPacket() (net.PacketConn, error) {
	addr := s.Addr[len(transport.DNSCrypt+"://"):]
	p, err := reuseport.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s ServerDNSCrypt) ServePacket(p net.PacketConn) error {
	if UDPConn, ok := p.(*net.UDPConn); ok {
		err := s.dnsCryptServer.ServeUDP(UDPConn)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("fail conversion %T to *net.UDPConn", p)
}
