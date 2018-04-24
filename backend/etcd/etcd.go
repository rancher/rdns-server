package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/niusmallnan/rdns-server/model"
	"github.com/sirupsen/logrus"
)

const (
	BackendName  = "etcd"
	ValueHostKey = "host"
	DefaultTTL   = "240h"

	maxSlugHashTimes = 100
)

type BackendOperator struct {
	kapi     client.KeysAPI
	prePath  string
	duration time.Duration
	baseFqdn string
}

func NewEtcdBackend(endpoints []string, prePath string, baseFqdn string) (*BackendOperator, error) {
	logrus.Debugf("Etcd init...")
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

	duration, err := time.ParseDuration(DefaultTTL)
	if err != nil {
		return nil, err
	}

	return &BackendOperator{kapi, prePath, duration, baseFqdn}, nil
}

func (e *BackendOperator) path(domainName string) string {
	return e.prePath + convertToPath(domainName)
}

func (e *BackendOperator) lookupHosts(path string) (hosts []string, err error) {
	opts := &client.GetOptions{Recursive: true}
	resp, err := e.kapi.Get(context.Background(), path, opts)
	if err != nil {
		return hosts, err
	}
	for _, n := range resp.Node.Nodes {
		v, err := convertToMap(n.Value)
		if err != nil {
			return hosts, err
		}
		hosts = append(hosts, v[ValueHostKey])
	}

	return hosts, nil
}

func (e *BackendOperator) refreshExpiration(path string, dopts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("Etcd: refresh dir TTL: %s", path)
	opts := &client.SetOptions{TTL: e.duration, Dir: true, PrevExist: client.PrevExist}
	resp, err := e.kapi.Set(context.Background(), path, "", opts)
	if err != nil {
		return d, err
	}

	curHosts, err := e.lookupHosts(path)
	if err != nil {
		return d, err
	}

	d.Fqdn = dopts.Fqdn
	d.Hosts = curHosts
	d.Expiration = resp.Node.Expiration
	return d, err
}

func (e *BackendOperator) set(path string, dopts *model.DomainOptions, exist bool) (d model.Domain, err error) {
	opts := &client.SetOptions{TTL: e.duration, Dir: true}
	if exist {
		opts.PrevExist = client.PrevExist
	}
	logrus.Debugf("Etcd: make a dir: %s", path)
	resp, err := e.kapi.Set(context.Background(), path, "", opts)
	if err != nil {
		return d, err
	}

	// get current hosts
	curHosts, err := e.lookupHosts(path)
	if err != nil {
		return d, err
	}

	// set key value
	newHostsMap := sliceToMap(dopts.Hosts)
	logrus.Debugf("Got new hosts map: %v", newHostsMap)
	oldHostsMap := sliceToMap(curHosts)
	logrus.Debugf("Got old hosts map: %v", oldHostsMap)
	for oldh := range oldHostsMap {
		if _, ok := newHostsMap[oldh]; !ok {
			key := fmt.Sprintf("%s/%s", path, formatKey(oldh))
			logrus.Debugf("Etcd: delete a key/value: %s:%s", key, formatValue(oldh))
			_, err := e.kapi.Delete(context.Background(), key, nil)
			if err != nil {
				return d, err
			}
		}
	}
	for newh := range newHostsMap {
		if _, ok := oldHostsMap[newh]; !ok {
			key := fmt.Sprintf("%s/%s", path, formatKey(newh))
			logrus.Debugf("Etcd: set a key/value: %s:%s", key, formatValue(newh))
			_, err := e.kapi.Set(context.Background(), key, formatValue(newh), nil)
			if err != nil {
				return d, err
			}
		}
	}

	d.Fqdn = dopts.Fqdn
	d.Hosts = dopts.Hosts
	d.Expiration = resp.Node.Expiration
	logrus.Debugf("Finished to set a domain entry: %s", d.String())

	return d, nil
}

func (e *BackendOperator) Get(dopts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("Get in etcd: Got the domain options entry: %s", dopts.String())
	path := e.path(dopts.Fqdn)
	//opts := &client.GetOptions{Recursive: true}
	resp, err := e.kapi.Get(context.Background(), path, nil)
	if err != nil {
		return d, err
	}

	d.Fqdn = dopts.Fqdn
	d.Expiration = resp.Node.Expiration
	for _, n := range resp.Node.Nodes {
		if n.Dir {
			continue
		}
		v, err := convertToMap(n.Value)
		if err != nil {
			return d, err
		}
		d.Hosts = append(d.Hosts, v[ValueHostKey])
	}

	return d, nil
}

func (e *BackendOperator) Create(dopts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("Create in etcd: Got the domain options entry: %s", dopts.String())
	var path string
	for i := 0; i < maxSlugHashTimes; i++ {
		fqdn := fmt.Sprintf("%s.%s", generateSlug(), e.baseFqdn)
		path = e.path(fqdn)

		// check if this path exists and use this path if not exist
		_, err := e.kapi.Get(context.Background(), path, nil)
		if client.IsKeyNotFound(err) {
			dopts.Fqdn = fqdn
			break
		}
	}

	d, err = e.set(path, dopts, false)
	if err != nil {
		return d, err
	}

	return e.Get(dopts)
}

func (e *BackendOperator) Update(dopts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("Update in etcd: Got the domain options entry: %s", dopts.String())
	path := e.path(dopts.Fqdn)

	d, err = e.set(path, dopts, true)
	return d, err
}

func (e *BackendOperator) Renew(dopts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("Renew in etcd: Got the domain options entry: %s", dopts.String())
	path := e.path(dopts.Fqdn)

	d, err = e.refreshExpiration(path, dopts)
	return d, err
}

func (e *BackendOperator) Delete(dopts *model.DomainOptions) error {
	logrus.Debugf("Delete in etcd: Got the domain options entry: %s", dopts.String())
	path := e.path(dopts.Fqdn)

	opts := &client.DeleteOptions{Dir: true, Recursive: true}
	_, err := e.kapi.Delete(context.Background(), path, opts)
	return err
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
// source data: {"host":"1.1.1.1"}
func convertToMap(value string) (map[string]string, error) {
	var v map[string]string
	err := json.Unmarshal([]byte(value), &v)
	return v, err
}

// formatValue
// 1.1.1.1 => {"host": "1.1.1.1"}
func formatValue(value string) string {
	return fmt.Sprintf("{\"%s\":\"%s\"}", ValueHostKey, value)
}

// formatKey
// 1.1.1.1 => 1_1_1_1
func formatKey(key string) string {
	return strings.Replace(key, ".", "_", -1)
}

func sliceToMap(ss []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range ss {
		m[s] = true
	}
	return m
}

// generateSlug will generate a random slug to be used as shorten link.
func generateSlug() string {
	// It doesn't exist! Generate a new slug for it
	// From: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
	var chars = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, 6)
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}

	return string(s)
}
