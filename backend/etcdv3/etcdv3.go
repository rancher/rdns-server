package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rancher/rdns-server/model"
	"github.com/rancher/rdns-server/util"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	Name             = "etcdv3"
	typeA            = "A"
	typeTXT          = "TXT"
	typeToken        = "TOKEN"
	typeFrozen       = "FROZEN"
	tokenPath        = "/tokenv3"
	frozenPath       = "/frozenv3"
	maxSlugHashTimes = 100
	tokenLength      = 32
	slugLength       = 6
	operationTimeout = 100 * time.Millisecond
)

type Backend struct {
	Domain    string
	Prefix    string
	FrozenTTL time.Duration
	TTL       time.Duration

	C *clientv3.Client
}

func NewBackend() (*Backend, error) {
	cfg := clientv3.Config{
		Endpoints:   strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","),
		DialTimeout: 5 * time.Second,
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	ttl, err := time.ParseDuration(os.Getenv("TTL"))
	if err != nil {
		return nil, err
	}
	frozen, err := time.ParseDuration(os.Getenv("FROZEN"))
	if err != nil {
		return nil, err
	}

	return &Backend{
		Domain:    os.Getenv("DOMAIN"),
		Prefix:    os.Getenv("ETCD_PREFIX_PATH"),
		FrozenTTL: frozen,
		TTL:       ttl,
		C:         c,
	}, nil
}

func (b *Backend) GetName() string {
	return Name
}

func (b *Backend) GetZone() string {
	return b.Domain
}

func (b *Backend) Get(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("get %s record for domain options: %s", typeA, opts.String())

	path := getPath(b.Prefix, opts.Fqdn)

	kvs, err := b.lookupKeys(path)
	if err != nil {
		return d, err
	}

	if len(kvs) <= 0 {
		return d, errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	subs := make(map[string][]string, 0)
	hosts := make([]string, 0)

	for _, v := range kvs {
		k := string(v.Key)
		prefix := findSubPrefix(k, path)

		m, err := unmarshalToMap(v.Value)
		if err != nil {
			return d, err
		}

		isText := false
		if _, ok := m["text"]; ok {
			isText = true
		}

		if prefix != "" && !strings.Contains(prefix, "_") && !isText {
			subs[prefix] = make([]string, 0)
			continue
		}

		if m["host"] == "" {
			continue
		}

		hosts = append(hosts, m["host"])
	}

	lease, err := b.getLease(kvs[0].Lease)
	if err != nil {
		return d, err
	}

	for k := range subs {
		n := fmt.Sprintf("%s.%s", k, opts.Fqdn)
		p := getPath(b.Prefix, n)

		kvs, err := b.lookupKeys(p)
		if err != nil {
			return d, err
		}

		ss := make([]string, 0)
		for _, v := range kvs {
			m, err := unmarshalToMap(v.Value)
			if err != nil {
				return d, err
			}
			ss = append(ss, m["host"])
		}

		subs[k] = ss
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = hosts
	d.SubDomain = subs
	d.Expiration = getExpiration(lease.TTL)

	return d, nil
}

func (b *Backend) Set(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("set %s record for domain options: %s", typeA, opts.String())

	var path, slug string
	for i := 0; i < maxSlugHashTimes; i++ {
		slug = generateSlug()

		if b.checkSlugName(slug) {
			logrus.Debugf(errExistSlug, slug)
			continue
		}

		fqdn := fmt.Sprintf("%s.%s", slug, b.Domain)
		path = getPath(b.Prefix, fqdn)

		if !b.checkPathExist(path) {
			opts.Fqdn = fqdn
			break
		}
	}

	d, err = b.setRecord(path, opts, false)
	if err != nil {
		return d, err
	}

	if err := b.lockSlugName(opts.Fqdn, slug, false); err != nil {
		return d, err
	}

	return b.Get(opts)
}

func (b *Backend) Update(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("update %s record for domain options: %s", typeA, opts.String())

	path := getPath(b.Prefix, opts.Fqdn)

	kvs, err := b.lookupKeys(path)
	if err != nil {
		return d, errors.Wrapf(err, errEmptyRecord, typeA, opts.Fqdn)
	}

	if len(kvs) <= 0 {
		return d, errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	if _, err = b.setRecord(path, opts, true); err != nil {
		return d, err
	}

	d, err = b.Get(opts)
	if err != nil {
		return d, err
	}

	return d, b.lockSlugName(opts.Fqdn, findSlugWithZone(opts.Fqdn, b.Domain), true)
}

func (b *Backend) Delete(opts *model.DomainOptions) error {
	logrus.Debugf("delete %s record for domain options: %s", typeA, opts.String())

	d, err := b.Get(opts)
	if err != nil {
		return err
	}

	path := getPath(b.Prefix, opts.Fqdn)

	kvs, err := b.lookupKeys(path)
	if err != nil {
		return err
	}

	for _, v := range kvs {
		k := string(v.Key)
		prefix := findSubPrefix(k, path)
		path := getPath(b.Prefix, fmt.Sprintf("%s.%s", prefix, opts.Fqdn))

		if prefix != "" && strings.Contains(prefix, "_") {
			ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
			_, err := b.C.Delete(ctx, path)
			cancel()
			if err != nil {
				return errors.Wrapf(err, errDeleteRecord, typeA, opts.Fqdn)
			}
		}

	}
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Delete(ctx, path); err != nil {
		return errors.Wrapf(err, errDeleteRecord, typeA, opts.Fqdn)
	}
	for prefix := range d.SubDomain {
		path := getPath(b.Prefix, fmt.Sprintf("%s.%s", prefix, opts.Fqdn))

		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		_, err := b.C.Delete(ctx, path, clientv3.WithPrefix())
		cancel()
		if err != nil {
			return errors.Wrapf(err, errDeleteRecord, typeA, opts.Fqdn)
		}
	}

	return nil
}

func (b *Backend) Renew(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("renew %s record for domain options: %s", typeA, opts.String())

	path := getPath(b.Prefix, opts.Fqdn)

	leaseID, leaseTTL, err := b.setToken(opts, true)
	if err != nil {
		return d, err
	}

	_, leaseTTL, err = b.keepaliveOnce(leaseID)
	if err != nil {
		return d, err
	}

	kvs, err := b.lookupKeys(path)
	if err != nil {
		return d, err
	}

	if len(kvs) <= 0 {
		return d, errors.Errorf(errEmptyRecord, typeA, opts.String())
	}

	subs := make(map[string][]string, 0)
	hosts := make([]string, 0)

	for _, v := range kvs {
		k := string(v.Key)
		prefix := findSubPrefix(k, path)

		m, err := unmarshalToMap(v.Value)
		if err != nil {
			return d, err
		}

		isText := false
		if _, ok := m["text"]; ok {
			isText = true
		}

		if prefix != "" && !strings.Contains(prefix, "_") && !isText {
			subs[prefix] = make([]string, 0)
			continue
		}

		hosts = append(hosts, m["host"])
	}

	for k := range subs {
		n := fmt.Sprintf("%s.%s", k, opts.Fqdn)
		p := getPath(b.Prefix, n)

		kvs, err := b.lookupKeys(p)
		if err != nil {
			return d, err
		}

		ss := make([]string, 0)
		for _, v := range kvs {
			m, err := unmarshalToMap(v.Value)
			if err != nil {
				return d, err
			}
			ss = append(ss, m["host"])
		}

		subs[k] = ss
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = hosts
	d.SubDomain = subs
	d.Expiration = getExpiration(leaseTTL)

	return d, nil
}

func (b *Backend) SetText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("set %s record for domain options: %s", typeTXT, opts.String())

	if len(strings.Split(opts.Fqdn, "."))-len(strings.Split(b.Domain, ".")) <= 1 {
		return d, errors.Errorf(errNotValidFqdn, opts.Fqdn)
	}

	path := getPath(b.Prefix, opts.Fqdn)
	slug := findSlugWithZone(opts.Fqdn, b.Domain)
	base := fmt.Sprintf("%s.%s", slug, b.Domain)

	leaseID, _, err := b.setToken(&model.DomainOptions{Fqdn: base}, true)
	if err != nil {
		return d, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Put(ctx, path, formatTextValue(opts.Text), clientv3.WithLease(clientv3.LeaseID(leaseID))); err != nil {
		return d, errors.Wrapf(err, errSetRecordWithLease, typeTXT, d.Fqdn)
	}

	return b.GetText(opts)
}

func (b *Backend) GetText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("get %s record for domain options: %s", typeTXT, opts.String())

	if len(strings.Split(opts.Fqdn, "."))-len(strings.Split(b.Domain, ".")) <= 1 {
		return d, errors.Errorf(errNotValidFqdn, opts.Fqdn)
	}

	path := getPath(b.Prefix, opts.Fqdn)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	resp, err := b.C.Get(ctx, path)
	if err != nil {
		return d, errors.Wrapf(err, errEmptyRecord, typeTXT, opts.Fqdn)
	}

	if resp.Count <= 0 {
		return d, errors.Errorf(errEmptyRecord, typeTXT, opts.Fqdn)
	}

	lease, err := b.getLease(resp.Kvs[0].Lease)
	if err != nil {
		return d, err
	}

	m, err := unmarshalToMap(resp.Kvs[0].Value)
	if err != nil {
		return d, err
	}

	if _, ok := m["text"]; ok {
		d.Text = m["text"]
	}

	d.Fqdn = opts.Fqdn
	d.Expiration = getExpiration(lease.TTL)

	return d, nil
}

func (b *Backend) UpdateText(opts *model.DomainOptions) (d model.Domain, err error) {
	logrus.Debugf("update %s record for domain options: %s", typeTXT, opts.String())

	if len(strings.Split(opts.Fqdn, "."))-len(strings.Split(b.Domain, ".")) <= 1 {
		return d, errors.Errorf(errNotValidFqdn, opts.Fqdn)
	}

	if _, err := b.GetText(opts); err != nil {
		return d, errors.Errorf(errEmptyRecord, typeTXT, opts.Fqdn)
	}

	path := getPath(b.Prefix, opts.Fqdn)
	slug := findSlugWithZone(opts.Fqdn, b.Domain)
	base := fmt.Sprintf("%s.%s", slug, b.Domain)

	leaseID, _, err := b.setToken(&model.DomainOptions{Fqdn: base}, true)
	if err != nil {
		return d, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Put(ctx, path, formatTextValue(opts.Text), clientv3.WithLease(clientv3.LeaseID(leaseID)), clientv3.WithPrevKV()); err != nil {
		return d, errors.Wrapf(err, errSetRecordWithLease, typeTXT, d.Fqdn)
	}

	return b.GetText(opts)
}

func (b *Backend) DeleteText(opts *model.DomainOptions) error {
	logrus.Debugf("delete %s record for domain options: %s", typeTXT, opts.String())

	path := getPath(b.Prefix, opts.Fqdn)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Delete(ctx, path); err != nil {
		return errors.Wrapf(err, errDeleteRecord, typeTXT, opts.Fqdn)
	}

	return nil
}

func (b *Backend) GetToken(fqdn string) (string, error) {
	logrus.Debugf("get %s record for fqdn: %s", typeToken, fqdn)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	path := getTokenPath(fqdn)

	resp, err := b.C.Get(ctx, path)
	if err != nil {
		return "", err
	}

	if resp.Count <= 0 {
		return "", errors.Errorf(errEmptyRecord, typeToken, fqdn)
	}

	if resp.Count > 1 {
		return "", errors.Errorf(errMultiRecords, typeToken, fqdn)
	}

	return string(resp.Kvs[0].Value), nil
}

func (b *Backend) GetTokenCount() (int64, error) {
	logrus.Debugf("get %s record count", typeToken)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	resp, err := b.C.Get(ctx, tokenPath, clientv3.WithPrefix())
	if err != nil {
		if err == rpctypes.ErrKeyNotFound {

			return 0, nil
		}
		return 0, err
	}

	return resp.Count, nil
}

func (b *Backend) MigrateFrozen(opts *model.MigrateFrozen) error {
	return nil
}

func (b *Backend) MigrateToken(opts *model.MigrateToken) error {
	return nil
}

func (b *Backend) MigrateRecord(opts *model.MigrateRecord) error {
	return nil
}

func (b *Backend) setRecord(path string, opts *model.DomainOptions, exist bool) (d model.Domain, err error) {
	leaseID, leaseTTL, err := b.setToken(opts, exist)
	if err != nil {
		return d, err
	}

	if !exist {
		// make sure domain record is exist, although no hosts value
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel()

		_, err := b.C.Put(ctx, path, formatValue(""), clientv3.WithLease(clientv3.LeaseID(leaseID)))
		if err != nil {
			return d, err
		}
	}

	kvs, err := b.lookupKeys(path)
	if err != nil {
		return d, err
	}

	subs := make(map[string][]string, 0)
	hosts := make([]string, 0)

	for _, v := range kvs {
		k := string(v.Key)
		prefix := findSubPrefix(k, path)

		m, err := unmarshalToMap(v.Value)
		if err != nil {
			return d, err
		}

		isText := false
		if _, ok := m["text"]; ok {
			isText = true
		}

		if prefix != "" && !strings.Contains(prefix, "_") && !isText {
			subs[prefix] = make([]string, 0)
			continue
		}

		hosts = append(hosts, m["host"])
	}

	for k := range subs {
		n := fmt.Sprintf("%s.%s", k, opts.Fqdn)
		p := getPath(b.Prefix, n)

		kvs, err := b.lookupKeys(p)
		if err != nil {
			return d, err
		}

		ss := make([]string, 0)
		for _, v := range kvs {
			m, err := unmarshalToMap(v.Value)
			if err != nil {
				return d, err
			}
			ss = append(ss, m["host"])
		}

		subs[k] = ss
	}

	if err := b.syncRecords(opts.Hosts, hosts, path, clientv3.LeaseID(leaseID)); err != nil {
		return d, errors.Wrapf(err, errSyncRecords, typeA, opts.Fqdn)
	}

	if err := b.setSubRecords(opts, subs, leaseID); err != nil {
		return d, errors.Wrapf(err, errSetSubRecordsWithLease, typeA, opts.Fqdn)
	}

	d.Fqdn = opts.Fqdn
	d.Hosts = opts.Hosts
	d.SubDomain = opts.SubDomain
	d.Expiration = getExpiration(leaseTTL)

	return d, err
}

func (b *Backend) setSubRecords(opts *model.DomainOptions, origins map[string][]string, leaseID int64) error {
	for prefix := range origins {
		if _, ok := opts.SubDomain[prefix]; !ok {
			path := getPath(b.Prefix, fmt.Sprintf("%s.%s", prefix, opts.Fqdn))
			ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
			_, err := b.C.Delete(ctx, path, clientv3.WithPrefix())
			cancel()
			if err != nil {
				return err
			}
		}
	}

	for prefix, values := range opts.SubDomain {
		path := getPath(b.Prefix, fmt.Sprintf("%s.%s", prefix, opts.Fqdn))

		kvs, err := b.lookupKeys(path)
		if err != nil {
			return err
		}

		hosts := make([]string, 0)
		for _, v := range kvs {
			m, err := unmarshalToMap(v.Value)
			if err != nil {
				return err
			}
			hosts = append(hosts, m["host"])
		}

		if err := b.syncRecords(values, hosts, path, clientv3.LeaseID(leaseID)); err != nil {
			return errors.Wrapf(err, errSyncSubRecords, typeA, opts.Fqdn)
		}
	}

	return nil
}

func (b *Backend) syncRecords(new, old []string, path string, leaseID clientv3.LeaseID) error {
	left := sliceToMap(new)
	right := sliceToMap(old)

	for r := range right {
		if _, ok := left[r]; !ok {
			key := fmt.Sprintf("%s/%s", path, formatKey(r))
			ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
			_, err := b.C.Delete(ctx, key)
			cancel()
			if err != nil {
				return err
			}
		}
	}

	for l := range left {
		if _, ok := right[l]; !ok {
			key := fmt.Sprintf("%s/%s", path, formatKey(l))
			ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
			_, err := b.C.Put(ctx, key, formatValue(l), clientv3.WithLease(leaseID))
			cancel()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Backend) setToken(opts *model.DomainOptions, exist bool) (int64, int64, error) {
	logrus.Debugf("set %s for fqdn: %s", typeToken, opts.String())

	var token string
	var leaseID, leaseTTL int64

	path := getTokenPath(opts.Fqdn)

	if exist {
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		defer cancel()

		resp, err := b.C.Get(ctx, path)
		if err != nil {
			return 0, -1, errors.Wrapf(err, errEmptyRecord, typeToken, opts.Fqdn)
		}

		if resp.Count <= 0 {
			return 0, -1, errors.Errorf(errEmptyRecord, typeToken, opts.Fqdn)
		}

		token = string(resp.Kvs[0].Value)

		lease, err := b.getLease(resp.Kvs[0].Lease)
		if err != nil {
			return 0, -1, err
		}

		leaseID = int64(lease.ID)
		leaseTTL = lease.TTL
	} else {
		token = util.RandStringWithAll(tokenLength)

		id, ttl, err := b.grantLease(int64(b.TTL.Seconds()))
		if err != nil {
			return 0, -1, err
		}

		leaseID = id
		leaseTTL = ttl
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Put(ctx, path, token, clientv3.WithLease(clientv3.LeaseID(leaseID))); err != nil {
		return 0, -1, errors.Wrapf(err, errSetRecordWithLease, typeToken, opts.Fqdn)
	}

	return leaseID, leaseTTL, nil
}

func (b *Backend) lockSlugName(fqdn, slug string, exist bool) error {
	logrus.Debugf("lock slug name: %s", fqdn)

	path := fmt.Sprintf("%s%s/%s", b.Prefix, frozenPath, slug)

	var leaseID int64
	if exist {
		kvs, err := b.lookupKeys(path)
		if err != nil {
			return err
		}

		if len(kvs) <= 0 {
			return errors.Errorf(errEmptyRecord, typeA, fqdn)
		}

		leaseID = kvs[0].Lease
	} else {
		id, _, err := b.grantLease(int64(b.FrozenTTL.Seconds()))
		if err != nil {
			return err
		}

		leaseID = id
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	if _, err := b.C.Put(ctx, path, "", clientv3.WithLease(clientv3.LeaseID(leaseID))); err != nil {
		return errors.Wrapf(err, errSetRecordWithLease, typeFrozen, path)
	}

	return nil
}

func (b *Backend) lookupKeys(path string) ([]*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	resp, err := b.C.Get(ctx, path, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, errLookupRecords, typeA, path)
	}

	kvs := make([]*mvccpb.KeyValue, 0)
	for _, v := range resp.Kvs {
		if len(v.Value) > 0 {
			m, err := unmarshalToMap(v.Value)
			if err != nil {
				continue
			}
			if _, ok := m["text"]; ok {
				continue
			}
		} else {
			v.Value = []byte("")
		}
		kvs = append(kvs, v)
	}

	return kvs, nil
}

func (b *Backend) getLease(id int64) (*clientv3.LeaseTimeToLiveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	lease, err := b.C.TimeToLive(ctx, clientv3.LeaseID(id))
	if err != nil {
		return nil, err
	}

	return lease, nil
}

func (b *Backend) grantLease(ttl int64) (int64, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	lease, err := b.C.Grant(ctx, ttl)
	if err != nil {
		return 0, -1, errors.Errorf(errGrantLease)
	}

	return int64(lease.ID), lease.TTL, nil
}

func (b *Backend) keepaliveOnce(id int64) (int64, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	keepalive, err := b.C.KeepAliveOnce(ctx, clientv3.LeaseID(id))
	if err != nil {
		return 0, -1, errors.Errorf(errKeepaliveOnce, id)
	}

	return int64(keepalive.ID), keepalive.TTL, nil
}

// Used to check whether fqdn can be used.
// e.g. sample.lb.rancher.cloud => /frozenv3/sample
// e.g. if /frozenv3/sample is exist that fqdn can not be used
func (b *Backend) checkSlugName(slug string) bool {
	path := fmt.Sprintf("%s%s/%s", b.Prefix, frozenPath, slug)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	resp, err := b.C.Get(ctx, path)
	if err != nil || resp.Count <= 0 {
		return false
	}

	return true
}

// Used to check whether path exist.
func (b *Backend) checkPathExist(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	resp, err := b.C.Get(ctx, path)
	if err != nil || resp.Count <= 0 {
		return false
	}

	return true
}

// Used to get a path as etcd preferred
// e.g. sample.lb.rancher.cloud => /rdnsv3/cloud/rancher/lb/sample
func getPath(path, fqdn string) string {
	return path + convertToPath(fqdn)
}

// Used to convert domain to a path as etcd preferred
// e.g. sample.lb.rancher.cloud => /cloud/rancher/lb/sample
func convertToPath(domain string) string {
	ss := strings.Split(domain, ".")
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
	return "/" + strings.Join(ss, "/")
}

// Used to get a token path as etcd preferred
// e.g. sample.lb.rancher.cloud => /tokenv3/sample_lb_rancher_cloud
func getTokenPath(fqdn string) string {
	return fmt.Sprintf("%s/%s", tokenPath, formatKey(fqdn))
}

// Used to format a key as etcd preferred
// e.g. 1.1.1.1 => 1_1_1_1
// e.g. sample.lb.rancher.cloud => sample_lb_rancher_cloud
func formatKey(key string) string {
	return strings.Replace(key, ".", "_", -1)
}

// Used to format a A value as dns preferred
// e.g. 1.1.1.1 => {"host": "1.1.1.1"}
func formatValue(value string) string {
	return fmt.Sprintf("{\"host\":\"%s\"}", value)
}

// Used to format a txt value as dns preferred
// e.g. abc => {"text": "abc"}
func formatTextValue(value string) string {
	return fmt.Sprintf("{\"text\":\"%s\"}", value)
}

// Used to generate a random slug
func generateSlug() string {
	return util.RandStringWithSmall(slugLength)
}

// Used to find slug name
// e.g. yyyy.xxxx.qrn7oq.lb.rancher.cloud => qrn7oq
func findSlugWithZone(fqdn, domain string) string {
	n := len(strings.Split(fqdn, ".")) - (len(strings.Split(domain, ".")))
	ss := strings.SplitN(fqdn, ".", n)
	return strings.Split(ss[len(ss)-1], ".")[0]
}

// Used to find sub domain prefix
// e.g. /rdnsv3/cloud/rancher/lb/jc1af/x1/1_1_1_1 => x1
func findSubPrefix(path, base string) string {
	if path == base {
		return ""
	}
	ss := strings.Split(path, base)
	prefix := strings.Split(ss[1], "/")
	return prefix[1]
}

// Used to get expiration time which etcd preferred
func getExpiration(ttl int64) *time.Time {
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", ttl))
	e := time.Now().Add(duration)
	return &e
}

func unmarshalToMap(b []byte) (map[string]string, error) {
	var v map[string]string
	err := json.Unmarshal(b, &v)
	return v, err
}

func sliceToMap(ss []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range ss {
		m[s] = true
	}
	return m
}
