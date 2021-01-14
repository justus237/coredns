package dnsserver

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

type testPlugin struct{}

func (tp testPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return 0, nil
}

func (tp testPlugin) Name() string { return "testplugin" }

func testConfig(transport string, p plugin.Handler) *Config {
	c := &Config{
		Zone:        "example.com.",
		Transport:   transport,
		ListenHosts: []string{"127.0.0.1"},
		Port:        "53",
		Debug:       false,
	}

	c.AddPlugin(func(next plugin.Handler) plugin.Handler { return p })
	return c
}

func TestNewServer(t *testing.T) {
	_, err := NewServer("127.0.0.1:53", []*Config{testConfig("dns", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}

	_, err = NewServergRPC("127.0.0.1:53", []*Config{testConfig("grpc", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServergRPC, got %s", err)
	}

	_, err = NewServerTLS("127.0.0.1:53", []*Config{testConfig("tls", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServerTLS, got %s", err)
	}

	_, err = NewServerQUIC("127.0.0.1:554", []*Config{testConfig("quic", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServerQUIC, got %s", err)
	}

}

func TestNewServerDNSCrypt(t *testing.T) {

	config := []byte(`provider_name: 2.dnscrypt-cert.example.org
public_key: C6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
private_key: B0B6DBF5BA3DA876992C092559AE044C0AFF30BF6F8C76496090E2881E4F479DC6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
resolver_secret: D7CB5AD6F0C4CDFEDD58541C95EED5030A0E01B8FFDD953D9B64D5B8ACA83820
resolver_public: 46F5E9EE56788B7272946FF5A355AE80D0F2574E4F698EB5EDE8D7290DC7B00F
es_version: 1
certificate_ttl: 0s`)

	dnscryptCfg := testConfig("dnscrypt", testPlugin{})
	err := yaml.Unmarshal(config, &dnscryptCfg.DNSCryptConfig)
	if err != nil {
		t.Fatalf("Expected no error for unmarshall test config data, got %s", err)
	}

	_, err = NewServerDNSCrypt("127.0.0.1:5443", []*Config{dnscryptCfg})
	if err != nil {
		t.Errorf("Expected no error for NewServerDNSCrypt, got %s", err)
	}
}

func BenchmarkCoreServeDNS(b *testing.B) {
	s, err := NewServer("127.0.0.1:53", []*Config{testConfig("dns", testPlugin{})})
	if err != nil {
		b.Errorf("Expected no error for NewServer, got %s", err)
	}

	ctx := context.TODO()
	w := &test.ResponseWriter{}
	m := new(dns.Msg)
	m.SetQuestion("aaa.example.com.", dns.TypeTXT)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeDNS(ctx, w, m)
	}
}
