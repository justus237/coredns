package test

import (
	"net"
	"testing"
	"time"

	"github.com/ameshkov/dnscrypt/v2"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

const (
	DnsCryptYamlConfig = `provider_name: 2.dnscrypt-cert.example.org
			public_key: C6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
			private_key: B0B6DBF5BA3DA876992C092559AE044C0AFF30BF6F8C76496090E2881E4F479DC6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
			resolver_secret: D7CB5AD6F0C4CDFEDD58541C95EED5030A0E01B8FFDD953D9B64D5B8ACA83820
			resolver_public: 46F5E9EE56788B7272946FF5A355AE80D0F2574E4F698EB5EDE8D7290DC7B00F
			es_version: 1
			certificate_ttl: 0s`
	tstCorefile = `dnscrypt://127.0.0.1:5443 {
		dnscrypt {
			` + DnsCryptYamlConfig + `
		}
	}`
	tstStamp = "sdns://AQcAAAAAAAAADjEyNy4wLjAuMTo1NDQzIMa_Z8yciMw-qnV30vymw3psTtucVI54nv5lu3wEEHN7GzIuZG5zY3J5cHQtY2VydC5leGFtcGxlLm9yZw"
)

func runInstance(t *testing.T) {
	g, _, _, err := CoreDNSServerAndPorts(tstCorefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}
	defer g.Stop()
}

func TestDnsCryptTcp(t *testing.T) {
	runInstance(t)

	client := getClient("tcp")
	resolver, err := client.Dial(tstStamp)
	if err != nil {
		t.Fatalf("Expected no errors, but: %s", err)
	}
	msg := getDnsMsg()

	reply, err := client.Exchange(&msg, resolver)
	assert.Nil(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, 1, len(reply.Answer))

	a, ok := reply.Answer[0].(*dns.A)
	assert.True(t, ok)
	assert.Equal(t, net.IPv4(8, 8, 8, 8).To4(), a.A.To4())

}

func TestDnsCryptUdp(t *testing.T) {
	runInstance(t)

	client := getClient("udp")
	resolver, err := client.Dial(tstStamp)
	if err != nil {
		t.Fatalf("Expected no errors, but: %s", err)
	}
	msg := getDnsMsg()

	reply, err := client.Exchange(&msg, resolver)
	assert.Nil(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, 1, len(reply.Answer))

	a, ok := reply.Answer[0].(*dns.A)
	assert.True(t, ok)
	assert.Equal(t, net.IPv4(8, 8, 8, 8).To4(), a.A.To4())

}

func getClient(network string) *dnscrypt.Client {
	return &dnscrypt.Client{
		Net:     network,
		Timeout: 5 * time.Second,
	}
}

func getDnsMsg() dns.Msg {
	req := dns.Msg{}
	req.Id = dns.Id()
	req.RecursionDesired = true
	req.Question = []dns.Question{
		{
			Name:   "google-public-dns-a.google.com.",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		},
	}
	return req
}
