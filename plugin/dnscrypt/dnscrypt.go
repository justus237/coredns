package dnscrypt

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"gopkg.in/yaml.v3"
)

const (
	pluginName = "dnscrypt"
)

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	err := parse(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}
	return nil
}

func parse(c *caddy.Controller) (err error) {
	config := dnsserver.GetConfig(c)

	if config.DNSCryptConfig != nil {
		return plugin.Error(pluginName, c.Errf("DNSCryptConfig already configured for this server instance"))
	}

	var coreFileTxt string
	var coreFileCfg []byte

	var yamlName string
	var yamlCfg []byte

	for i := 0; c.Next(); i++ {
		keyVal := c.Val()

		if keyVal == pluginName && c.NextArg() {
			name := c.Val()
			if strings.HasSuffix(name, "yaml") || strings.HasSuffix(name, "yml") {
				yamlName = name
			}
		}

		if strings.HasSuffix(keyVal, ":") && c.NextArg() {
			coreFileTxt += fmt.Sprintf("%s %s\n", keyVal, c.Val())
		}
	}

	if yamlName != "" {
		yamlCfg, err = ioutil.ReadFile(yamlName)
		if err != nil {
			return plugin.Error(pluginName, err)
		}
	}

	if coreFileTxt != "" {
		coreFileCfg = append(coreFileCfg, []byte(coreFileTxt)...)
	}

	err = yaml.Unmarshal(yamlCfg, &config.DNSCryptConfig)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	err = yaml.Unmarshal(coreFileCfg, &config.DNSCryptConfig)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	return
}
