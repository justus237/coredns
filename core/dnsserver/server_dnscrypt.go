package dnsserver

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/ameshkov/dnscrypt/v2"
	"github.com/coredns/coredns/plugin/pkg/reuseport"
	"github.com/coredns/coredns/plugin/pkg/transport"
)

type ServerDNSCrypt struct {
	*Server
	UDPConns       []*net.UDPConn
	udpWG          *sync.WaitGroup
	dnsCryptServer *dnscrypt.Server
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
		},
	}, nil
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
		for {
			err := s.dnsCryptServer.ServeUDP(UDPConn)
			if err != nil {
				return err
			}
		}
	}
	return fmt.Errorf("fail conversion %T to *net.UDPConn", p)
}
