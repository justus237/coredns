package tls

import (
	ctls "crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

// NextProtoDQ - During connection establishment, DNS/QUIC support is indicated
// by selecting the ALPN token "dq" in the crypto handshake.
// Current draft version: https://datatracker.ietf.org/doc/html/draft-ietf-dprive-dnsoquic-02
const NextProtoDQ = "doq-i02"

// nextProtosDQ are ALPNs for a DNS-over-QUIC server
var nextProtosDoQ = []string{
	NextProtoDQ, "doq-i00", "dq", "doq",
}

// nextProtosDoH are ALPNs for a DNS-over-HTTPS server
var nextProtosDoH = []string{
	"h2", "http/1.1",
}

const reloadPeriod = time.Minute

var log = clog.NewWithPlugin("tls")

func init() { plugin.Register("tls", setup) }

func setup(c *caddy.Controller) error {
	err := parseTLS(c)
	if err != nil {
		return plugin.Error("tls", err)
	}
	return nil
}

func setTLSDefaults(tls *ctls.Config) {
	tls.MinVersion = ctls.VersionTLS12
	tls.MaxVersion = ctls.VersionTLS13
	tls.CipherSuites = []uint16{
		ctls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		ctls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		ctls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		ctls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		ctls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		ctls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		ctls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		ctls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		ctls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	tls.PreferServerCipherSuites = true
}

func parseTLS(c *caddy.Controller) error {
	config := dnsserver.GetConfig(c)

	if config.TLSConfig != nil {
		return plugin.Error("tls", c.Errf("TLS already configured for this server instance"))
	}

	for c.Next() {
		args := c.RemainingArgs()
		//log.Infof("remaining args: %s\n", args)
		if len(args) == 0 || len(args)%2 != 0 {
			return plugin.Error("tls", c.ArgErr())
		}
		clientAuth := ctls.NoClientCert
		var sessionTicketKeysFiles []string
		for c.NextBlock() {
			switch c.Val() {
			case "client_auth":
				authTypeArgs := c.RemainingArgs()
				if len(authTypeArgs) != 1 {
					return c.ArgErr()
				}
				switch authTypeArgs[0] {
				case "nocert":
					clientAuth = ctls.NoClientCert
				case "request":
					clientAuth = ctls.RequestClientCert
				case "require":
					clientAuth = ctls.RequireAnyClientCert
				case "verify_if_given":
					clientAuth = ctls.VerifyClientCertIfGiven
				case "require_and_verify":
					clientAuth = ctls.RequireAndVerifyClientCert
				default:
					return c.Errf("unknown authentication type '%s'", authTypeArgs[0])
				}
			case "session_ticket_key":
				files := c.RemainingArgs()
				if len(files) == 0 {
					return c.ArgErr()
				}
				sessionTicketKeysFiles = append(sessionTicketKeysFiles, files...)
			default:
				return c.Errf("unknown option '%s'", c.Val())
			}
		}
		tls, err := newTlsConfigFromArgs(args)
		if err != nil {
			return err
		}
		tls.ClientAuth = clientAuth
		// NewTLSConfigFromArgs only sets RootCAs, so we need to let ClientCAs refer to it.
		tls.ClientCAs = tls.RootCAs

		setTLSDefaults(tls)

		// Load the default set of session tickets
		if sessionTicketKeysFiles != nil {
			err = loadSessionTickets(tls, sessionTicketKeysFiles)
			if err != nil {
				return err
			}
		}

		// DNS-over-QUIC config
		tlsDoQ := tls.Clone()
		tlsDoQ.NextProtos = nextProtosDoQ

		// DNS-over-HTTPS config
		tlsDoH := tls.Clone()
		tlsDoH.NextProtos = nextProtosDoH

		config.TLSConfig = tls
		config.TLSConfigQUIC = tlsDoQ
		config.TLSConfigHTTPS = tlsDoH

		// Add callbacks that would count TLS metrics
		config.TLSConfig.VerifyConnection = makeHandshakeMetrics("tls")
		config.TLSConfigQUIC.VerifyConnection = makeHandshakeMetrics("quic")
		config.TLSConfigHTTPS.VerifyConnection = makeHandshakeMetrics("https")

		// Schedule reloading of the TLS session tickets
		/*if sessionTicketKeysFiles != nil {
			go reloadSessionTickets(tls, sessionTicketKeysFiles)
			go reloadSessionTickets(tlsDoQ, sessionTicketKeysFiles)
			go reloadSessionTickets(tlsDoH, sessionTicketKeysFiles)
		}*/
	}
	return nil
}

func makeHandshakeMetrics(proto string) func(state ctls.ConnectionState) error {
	return func(state ctls.ConnectionState) error {
		didResume := "0"
		if state.DidResume {
			didResume = "1"
		}

		tlsVersion := "unknown"
		switch state.Version {
		case ctls.VersionTLS13:
			tlsVersion = "tls1.3"
		case ctls.VersionTLS12:
			tlsVersion = "tls1.2"
		case ctls.VersionTLS11:
			tlsVersion = "tls1.1"
		case ctls.VersionTLS10:
			tlsVersion = "tls1.0"
		}

		tlsHandshakeTotal.With(prometheus.Labels{
			"proto":            proto,
			"server_name":      state.ServerName,
			"tls_version":      tlsVersion,
			"did_resume":       didResume,
			"cipher_suite":     ctls.CipherSuiteName(state.CipherSuite),
			"negotiated_proto": state.NegotiatedProtocol,
		}).Inc()

		return nil
	}
}

func reloadSessionTickets(tls *ctls.Config, sessionTicketKeysFiles []string) {
	ticker := time.NewTicker(reloadPeriod)
	defer ticker.Stop()

	// sleep the first time -- we've already loaded the list
	time.Sleep(reloadPeriod)

	for t := range ticker.C {
		_ = t // we don't print the ticker time, so assign this `t` variable to underscore `_` to avoid error
		_ = loadSessionTickets(tls, sessionTicketKeysFiles)
	}
}

func loadSessionTickets(tls *ctls.Config, sessionTicketKeysFiles []string) error {
	var keys [][32]byte

	for _, file := range sessionTicketKeysFiles {
		b, err := ioutil.ReadFile(file)
		if err != nil || len(b) < 32 {
			tlsSessionTicketsRotateStatus.Set(0)
			clog.Errorf("failed to read session ticket from %s", file)
			return err
		}
		log.Infof("session ticket key file: %s\n", file)
		key := [32]byte{}
		copy(key[:], b[len(b)-32:])
		log.Infof("session ticket key: %x\n", key)
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		clog.Errorf("found no session tickets")
		return errors.New("no keys found")
	}

	tls.SetSessionTicketKeys(keys)
	tlsSessionTicketsRotateTime.SetToCurrentTime()
	tlsSessionTicketsRotateStatus.Set(1)
	return nil
}

func newTlsConfigFromArgs(args []string) (*ctls.Config, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("invalid number of tls arguments")
	}

	var certs []ctls.Certificate

	for i := 0; i < len(args); i += 2 {
		tlsArgs := args[i : i+2]

		cert, err := ctls.LoadX509KeyPair(tlsArgs[0], tlsArgs[1])
		if err != nil {
			return nil, fmt.Errorf("could not load TLS cert: %s", err)
		}
		certs = append(certs, cert)
	}

	return &ctls.Config{Certificates: certs}, nil
}
