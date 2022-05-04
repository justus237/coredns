package httpproxy

import (
	"net/http"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/lucas-clemente/quic-go/http3"
)

const pluginName = "h3server"

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

func parseHTTPProxy(c *caddy.Controller) error {
	//return plugin.Error(pluginName, c.Err("test"))
	config := dnsserver.GetConfig(c)
	if config.TLSConfigQUIC == nil {
		return plugin.Error(pluginName, c.Err("H3 server requires DoQ server to be configured (correctly) (i.e. need TLS config)"))
	}
	tlsConfig := config.TLSConfigQUIC.Clone()
	tlsConfig.NextProtos = nextProtosH3
	wwwDir := "/Users/justus/web-performance"
	handlerMux := http.NewServeMux()
	handlerMux.Handle("/", http.FileServer(http.Dir(wwwDir)))
	//http.Handle("/", http.FileServer(http.Dir(wwwDir)))
	/*tokenAcceptor := func(clientAddr net.Addr, token *quic.Token) bool {
		return true
	}*/
	server := http3.Server{
		Server: &http.Server{Handler: handlerMux, Addr: "localhost:4433"},
		//QuicConfig: &quic.Config{AcceptToken: tokenAcceptor},
	}
	server.TLSConfig = tlsConfig
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
