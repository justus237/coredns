package httpproxy

import (
	"net"
	"net/http"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

const pluginName = "h3server"

// maxQuicIdleTimeout - maximum QUIC idle timeout.
// Default value in quic-go is 30, but our internal tests show that
// a higher value works better for clients written with ngtcp2
const maxQuicIdleTimeout = 10 * time.Second //5 * time.Minute

var log = clog.NewWithPlugin(pluginName)

var nextProtosH3 = []string{
	"h3",
}

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	err := parseHTTPProxy(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}
	return nil
}

//based on https://github.com/zenazn/goji/blob/master/web/middleware/nocache.go
func cachingDisabledHTTPHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
		w.Header().Set("Expires", time.Unix(0, 0).Format(time.RFC1123))
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Accel-Expires", "0")
		h.ServeHTTP(w, r)
	})
}

func parseHTTPProxy(c *caddy.Controller) error {
	//return plugin.Error(pluginName, c.Err("test"))
	config := dnsserver.GetConfig(c)
	if config.TLSConfigQUIC == nil {
		return plugin.Error(pluginName, c.Err("H3 server requires DoQ server to be configured (correctly) (i.e. need TLS config)"))
	}
	tlsConfig := config.TLSConfigQUIC.Clone()
	tlsConfig.NextProtos = nil
	/*certs := tlsConfig.Certificates
	log.Infof("certs: %d\n", len(certs))
	cert := certs[0]
	log.Infof("cert: %x\n", cert.PrivateKey)*/
	var wwwDir string
	var hostAndPort string
	//get directory to serve
	for c.Next() {
		args := c.RemainingArgs()
		//only serve a single directory
		if len(args) != 2 {
			return plugin.Error(pluginName, c.ArgErr())
		}
		wwwDir = args[0]
		hostAndPort = args[1]
	}
	//TODO: implement gzip compression,
	//e.g. https://github.com/NYTimes/gziphandler
	handlerMux := http.NewServeMux()
	handlerMux.Handle("/", cachingDisabledHTTPHandler(http.FileServer(http.Dir(wwwDir))))
	//http.Handle("/", http.FileServer(http.Dir(wwwDir)))
	var customAcceptToken = func(clientAddr net.Addr, token *quic.Token) bool {
		/*log.Infof("token acceptor called for: %s\n", clientAddr.String())
		if token == nil {
			log.Infof("no token, rejecting and asking for retry\n")
			return false
		}
		log.Infof("token with remote addr: %s\n", token.RemoteAddr)*/
		return true
	}
	quicConf := &quic.Config{
		MaxIdleTimeout: maxQuicIdleTimeout,
		AcceptToken:    customAcceptToken,
		KeepAlive:      false,
		//StatelessResetKey: nil,
	}
	server := http3.Server{
		Server:     &http.Server{Handler: handlerMux, Addr: hostAndPort, TLSConfig: tlsConfig, IdleTimeout: 10 * time.Second},
		QuicConfig: quicConf,
	}
	server.SetKeepAlivesEnabled(false)
	//server.TLSConfig = tlsConfig
	go func() {
		err := server.ListenAndServe()
		plugin.Error(pluginName, err)
	}()
	return nil
	/*
		if config.HTTPProxy != nil {
			return plugin.Error(pluginName, c.Errf("HTTPProxy already configured for this server instance"))
		}

		for c.Next() {
			args := c.RemainingArgs()
			if len(args) != 1 {
				return plugin.Error(pluginName, c.ArgErr())
			}

			host, port, err := net.SplitHostPort(args[0])
			if err != nil {
				return c.ArgErr()
			}

			proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%s", host, port))
			if err != nil {
				return c.ArgErr()
			}

			reverseProxy := httputil.NewSingleHostReverseProxy(proxyURL)
			config.HTTPProxy = reverseProxy
		}

		return nil*/
}
