module github.com/rancher/rdns-server

require (
	github.com/aws/aws-sdk-go v1.20.4
	github.com/coredns/coredns v1.5.0
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/context v1.1.1
	github.com/gorilla/mux v1.7.2
	github.com/mholt/caddy v0.11.5
	github.com/miekg/dns v1.1.6
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20190618222545-ea8f1a30c443
	k8s.io/api v0.0.0-20190111032252-67edc246be36
	k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
)
