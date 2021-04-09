package httpproxy

import (
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const pluginName = "httpproxy"

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
	config := dnsserver.GetConfig(c)

	if config.HTTPProxy != nil {
		return plugin.Error("tls", c.Errf("HTTPProxy already configured for this server instance"))
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

	return nil
}
