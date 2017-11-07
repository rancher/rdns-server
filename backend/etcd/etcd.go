package etcd

import (
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

const ETCD_BACKEND = "etcd"

type EtcdBackend struct {
	kapi *client.KeysAPI
}

func NewEtcdBackend(endpoints []string) (*EtcdBackend, error) {
	cfg := client.Config{
		Endpoints: endpoints,
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	kapi := client.NewKeysAPI(c)

	return &EtcdBackend{kapi}, nil
}

func (e *EtcdBackend) Get(domain string) {

}

func (e *EtcdBackend) Create(domain string) {

}

func (e *EtcdBackend) Update(domain string) {

}

func (e *EtcdBackend) Delete(domain string) {

}

func (e *EtcdBackend) List(domain string) {

}

// convertToPath
// zhibo.test.rancher.local => /local/rancher/test/zhibo
func convertToPath(domain string) string {
	ss := strings.Split(domain, ".")

	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}

	return strings.Join(ss, "/")
}
