package rdns

import (
	"crypto/tls"
	"strconv"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	mwtls "github.com/coredns/coredns/plugin/pkg/tls"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	etcdcv3 "github.com/coreos/etcd/clientv3"
	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin("rdns")

func Setup(c *caddy.Controller) error {
	e, err := etcdParse(c)
	if err != nil {
		return plugin.Error("rdns", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		e.Next = next
		return e
	})

	return nil
}

func etcdParse(c *caddy.Controller) (*ETCD, error) {
	etc := ETCD{PathPrefix: "skydns"}
	var (
		tlsConfig *tls.Config
		err       error
		endpoints = []string{defaultEndpoint}
		username  string
		password  string
	)
	for c.Next() {
		etc.Zones = c.RemainingArgs()
		if len(etc.Zones) == 0 {
			etc.Zones = make([]string, len(c.ServerBlockKeys))
			copy(etc.Zones, c.ServerBlockKeys)
		}
		for i, str := range etc.Zones {
			etc.Zones[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "stubzones":
				// ignored, remove later.
			case "fallthrough":
				etc.Fall.SetZonesFromArgs(c.RemainingArgs())
			case "debug":
				/* it is a noop now */
			case "path":
				if !c.NextArg() {
					return &ETCD{}, c.ArgErr()
				}
				etc.PathPrefix = c.Val()
			case "endpoint":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return &ETCD{}, c.ArgErr()
				}
				endpoints = args
			case "upstream":
				// check args != 0 and error in the future
				c.RemainingArgs() // clear buffer
				etc.Upstream = upstream.New()
			case "tls": // cert key cacertfile
				args := c.RemainingArgs()
				tlsConfig, err = mwtls.NewTLSConfigFromArgs(args...)
				if err != nil {
					return &ETCD{}, err
				}
			case "credentials":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return &ETCD{}, c.ArgErr()
				}
				if len(args) != 2 {
					return &ETCD{}, c.Errf("credentials requires 2 arguments, username and password")
				}
				username, password = args[0], args[1]
			case "wildcardbound":
				if !c.NextArg() {
					return &ETCD{}, c.ArgErr()
				}
				v, err := strconv.ParseInt(c.Val(), 10, 8)
				if err != nil {
					return &ETCD{}, err
				}
				if v < 0 {
					return &ETCD{}, c.Errf("wildcardbound value can not be negative: %d", v)
				}
				etc.WildcardBound = int8(v)
			default:
				if c.Val() != "}" {
					return &ETCD{}, c.Errf("unknown property '%s'", c.Val())
				}
			}
		}
		client, err := newEtcdClient(endpoints, tlsConfig, username, password)
		if err != nil {
			return &ETCD{}, err
		}
		etc.Client = client
		etc.endpoints = endpoints

		return &etc, nil
	}
	return &ETCD{}, nil
}

func newEtcdClient(endpoints []string, cc *tls.Config, username, password string) (*etcdcv3.Client, error) {
	etcdCfg := etcdcv3.Config{
		Endpoints: endpoints,
		TLS:       cc,
	}
	if username != "" && password != "" {
		etcdCfg.Username = username
		etcdCfg.Password = password
	}
	cli, err := etcdcv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

const defaultEndpoint = "http://localhost:2379"
