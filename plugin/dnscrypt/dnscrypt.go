package dnscrypt

import (
	"io/ioutil"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const defaultName = "dnscrypt.yaml"

func init() { plugin.Register("dnscrypt", setup) }

func setup(c *caddy.Controller) error {
	err := parse(c)
	if err != nil {
		return plugin.Error("dnscrypt", err)
	}
	return nil
}

func parse(c *caddy.Controller) error {
	config := dnsserver.GetConfig(c)

	if config.DNSCryptConfig != nil {
		txt := "DNSCrypt already configured for this server instance"
		return plugin.Error("dnscrypt", c.Errf(txt))
	}

	args := c.Dispenser.RemainingArgs()

	filename := ""
	for _, arg := range args {
		yamlSuffix := strings.HasSuffix(arg, ".yaml")
		ymlSuffix := strings.HasSuffix(arg, ".yml")
		if ymlSuffix || yamlSuffix {
			filename = arg
		}
	}

	if filename == "" {
		filename = defaultName
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return plugin.Error("file", err)
	}

	config.DNSCryptConfig = data

	return nil
}
