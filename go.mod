module github.com/coredns/coredns

go 1.13

//replace github.com/lucas-clemente/quic-go => ./replacement_modules/github.com/lucas-clemente/quic-go@v0.21.1
//replace github.com/marten-seemann/qtls-go1-15 => ./replacement_modules/github.com/marten-seemann/qtls-go1-15@v0.1.4
//replace github.com/marten-seemann/qtls-go1-16 => ./replacement_modules/github.com/marten-seemann/qtls-go1-16@v0.1.3
//replace github.com/marten-seemann/qtls-go1-17 => ./replacement_modules/github.com/marten-seemann/qtls-go1-17@v0.1.0-beta.1.2

require (
	github.com/Azure/azure-sdk-for-go v40.6.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.3
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.1
	github.com/DataDog/datadog-go v3.5.0+incompatible // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/Shopify/sarama v1.21.0 // indirect
	github.com/ameshkov/dnscrypt/v2 v2.2.1
	github.com/apache/thrift v0.13.0 // indirect
	github.com/aws/aws-sdk-go v1.34.5
	github.com/caddyserver/caddy v1.0.5
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/dnstap/golang-dnstap v0.2.0
	github.com/farsightsec/golang-framestream v0.0.0-20190425193708-fa4b164d59b8
	github.com/golang/protobuf v1.4.2
	github.com/gophercloud/gophercloud v0.9.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/infobloxopen/go-trees v0.0.0-20190313150506-2af4e13f9062
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lucas-clemente/quic-go v0.21.1
	github.com/marten-seemann/qtls-go1-15 v0.1.4
	github.com/marten-seemann/qtls-go1-16 v0.1.3
	github.com/marten-seemann/qtls-go1-17 v0.1.0-beta.1.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/dns v1.1.40
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.10.0
	github.com/stretchr/testify v1.6.1
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200306183522-221f0cc107cb
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007
	google.golang.org/api v0.29.0
	google.golang.org/grpc v1.29.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.26.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/klog v1.0.0
)

replace github.com/miekg/dns => github.com/AdguardTeam/dns v1.1.36-0.20210422142401-68bfced682f6
