# DNSCrypt

## Name

*DNSCrypt* - allows you to configure the server certificates for the DNSCrypt servers.

## Description

Allows to specify the DNSCrypt configuration for the current server block.

## Syntax

~~~ txt
dnscrypt://.:5443 {
    dnscrypt dnscrypt.{yaml|yml}
}
~~~

or

~~~ txt
dnscrypt://.:5443 {
    dnscrypt {
        provider_name: 2.dnscrypt-cert.example.org
        ...
    }
}
~~~

Configuration from yaml file has a lower priority and can be overwritten by parameters from Corefile.

## Examples

~~~
provider_name: 2.dnscrypt-cert.example.org
public_key: C6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
private_key: B0B6DBF5BA3DA876992C092559AE044C0AFF30BF6F8C76496090E2881E4F479DC6BF67CC9C88CC3EAA7577D2FCA6C37A6C4EDB9C548E789EFE65BB7C0410737B
resolver_secret: D7CB5AD6F0C4CDFEDD58541C95EED5030A0E01B8FFDD953D9B64D5B8ACA83820
resolver_public: 46F5E9EE56788B7272946FF5A355AE80D0F2574E4F698EB5EDE8D7290DC7B00F
es_version: 1
certificate_ttl: 0s
~~~

* `provider_name` - DNSCrypt resolver name.
* `public_key`, `private_key` - keypair that is used by the DNSCrypt resolver to sign the certificate.
* `resolver_secret`, `resolver_public` - keypair that is used by the DNSCrypt resolver to encrypt and decrypt messages.
* `es_version` - crypto to use. Can be `1` (XSalsa20Poly1305) or `2` (XChacha20Poly1305).
* `certificate_ttl` - certificate time-to-live. By default it's set to `0` and in this case 1-year cert is generated. The certificate is generated on `dnscrypt` start-up and it will only be valid for the specified amount of time. You should periodically restart `dnscrypt` to rotate the cert.

To generate correct configuration or check the server, use this client:
https://github.com/ameshkov/dnscrypt

## Also See

[Official protocol homepage](https://dnscrypt.info/)

[DNSCrypt stamp calculator](https://dnscrypt.info/stamps)
