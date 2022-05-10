#!/bin/bash
#source: https://letsencrypt.org/de/docs/certificates-for-localhost/
# using the google chrome sample version from the source for the fingerprint works just fine as well
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

#source: https://github.com/GoogleChrome/samples/blob/gh-pages/quictransport/quic_transport_server.py#L44
# (perma link in case they move the file https://github.com/GoogleChrome/samples/blob/e16a665b10f055824a6c4b39b447fc255b03dec6/quictransport/quic_transport_server.py#L61)
echo "fingerprint:"
echo $(openssl x509 -pubkey -noout -in localhost.crt | openssl rsa -pubin -outform der | openssl dgst -sha256 -binary | base64)