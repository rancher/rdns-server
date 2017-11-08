package etcd

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

const (
	ETCD_BACKEND   = "etcd"
	VALUE_HOST_KEY = "host"
	DEFAULT_TTL    = "240h"
)

type EtcdBackend struct {
	kapi     *client.KeysAPI
	prePath  string
	duration time.Duration
}

func NewEtcdBackend(endpoints []string, prePath string) (*EtcdBackend, error) {
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

	duration, err := time.ParseDuration(DEFAULT_TTL)
	if err != nil {
		return nil, err
	}

	return &EtcdBackend{kapi, prePath, duration}, nil
}

func (e *EtcdBackend) path(domainName string) string {
	return e.prePath + convertToPath(domainName)
}

func (e *EtcdBackend) set(path string, dopts *DomainOptions, refresh bool) (d Domain, err error) {
	// mkdir/update a dir and set TTL
	opts := &client.SetOptions{TTL: e.duration, Dir: true}
	if refresh {
		opts.PrevExist = client.PrevExist
	}
	resp, err := kapi.Set(context.Background(), path, nil, opts)
	if err != nil {
		return d, err
	}

	// set key value
	for _, h := range dopts.Hosts {
		key := formatKey(h) + path
		resp, err := kapi.Set(context.Background(), key, h, nil)
		if err != nil {
			return d, err
		}
	}

	return d, nil
}

func (e *EtcdBackend) Get(dopts *DomainOptions) (d Domain, err error) {
	path := e.path(dopts.Fqdn)
	//opts := &client.GetOptions{Recursive: true}
	resp, err := e.kapi.Get(context.Background(), path, nil)
	if err != nil {
		return d, err
	}

	d.Expiration = resp.Node.Expiration
	for _, n := range resp.Node.Nodes {
		if n.Dir {
			continue
		}
		v, err := convertToMap(n.Value)
		if err != nil {
			return d, err
		}
		d.hosts = append(d.hosts, v[VALUE_HOST_KEY])
	}

	return d, nil
}

func (e *EtcdBackend) Create(dopts *DomainOptions) (d Domain, err error) {
	path := e.path(dopts.Fqdn)

	d, err = e.set(path, dopts, false)
	if err != nil {
		return d, err
	}

	return e.Get(d)
}

func (e *EtcdBackend) Update(dopts *DomainOptions) (d Domain, err error) {
	path := e.path(dopts.Fqdn)

	d, err = e.set(path, dopts, true)
	if err != nil {
		return d, err
	}

	return e.Get(d)
}

func (e *EtcdBackend) Renew(dopts *DomainOptions) (d Domain, err error) {
	return Update(dopts)
}

func (e *EtcdBackend) Delete(dopts *DomainOptions) error {
	path := e.path(dopts.Fqdn)

	opts := &client.DeleteOptions{Dir: true, Recursive: true}
	resp, err = e.kapi.Delete(context.Background(), path, opts)
	if err != nil {
		return err
	}

	return nil
}

// convertToPath
// zhibo.test.rancher.local => /local/rancher/test/zhibo
func convertToPath(domain string) string {
	ss := strings.Split(domain, ".")

	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}

	return "/" + strings.Join(ss, "/")
}

// convertToMap
// {"host":"1.1.1.1"}
func convertToMap(value string) (map[string]string, error) {
	var v map[string]string
	err := json.Unmarshal([]byte(value), &v)
	return v, err
}

// formatKey
// 1.1.1.1 => 1_1_1_1
func formatKey(key string) string {
	return strings.Replace(key, ".", "_", -1)
}
