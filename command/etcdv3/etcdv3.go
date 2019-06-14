package etcdv3

import (
	"net/http"
	"os"
	"strings"

	"github.com/rancher/rdns-server/backend"
	"github.com/rancher/rdns-server/backend/etcdv3"
	"github.com/rancher/rdns-server/metric"
	"github.com/rancher/rdns-server/service"

	"github.com/coredns/coredns/coremain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	// Plug in CoreDNS
	_ "github.com/coredns/coredns/core/plugin"
)

var (
	flags = map[string]map[string]string{
		"DOMAIN":           {"used to set etcd root domain.": "lb.rancher.cloud"},
		"ETCD_ENDPOINTS":   {"used to set etcd endpoints.": "http://127.0.0.1:2379"},
		"ETCD_PREFIX_PATH": {"used to set etcd prefix path.": "/rdnsv3"},
		"CORE_FILE":        {"used to set coredns file.": "/etc/coredns/config/corefile"},
		"TTL":              {"used to set record ttl.": "240h"},
	}
)

func Flags() []cli.Flag {
	fgs := make([]cli.Flag, 0)
	for key, value := range flags {
		for k, v := range value {
			f := cli.StringFlag{
				Name:   strings.ToLower(key),
				EnvVar: key,
				Usage:  k,
				Value:  v,
			}
			fgs = append(fgs, f)
		}
	}
	return fgs
}

func Action(c *cli.Context) error {
	if err := setEnvironments(c); err != nil {
		return errors.Wrapf(err, "failed to set environments")
	}

	b, err := setBackend()
	if err != nil {
		return err
	}

	defer func() {
		if err := b.C.Close(); err != nil {
			logrus.Fatalf("failed to close etcd-v3 client: %v", err)
		}
	}()

	done := make(chan struct{})

	go metric.StartMetricDaemon(done)

	go coremain.Run()

	go func() {
		if err := http.ListenAndServe(c.GlobalString("listen"), service.NewRouter()); err != nil {
			logrus.Error(err)
			done <- struct{}{}
		}
	}()

	<-done
	return nil
}

func setEnvironments(c *cli.Context) error {
	if c.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	for k := range flags {
		if err := os.Setenv(k, c.String(strings.ToLower(k))); err != nil {
			return err
		}
		if os.Getenv(k) == "" {
			return errors.Errorf("expected argument: %s", strings.ToLower(k))
		}
	}

	return os.Setenv("FROZEN", c.GlobalString("frozen"))
}

func setBackend() (*etcdv3.Backend, error) {
	b, err := etcdv3.NewBackend()
	if err != nil {
		return b, err
	}
	backend.SetBackend(b)

	return b, nil
}