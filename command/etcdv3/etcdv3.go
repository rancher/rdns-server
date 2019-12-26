package etcdv3

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/rancher/rdns-server/backend"
	"github.com/rancher/rdns-server/backend/etcdv3"
	"github.com/rancher/rdns-server/coredns"
	"github.com/rancher/rdns-server/metric"
	"github.com/rancher/rdns-server/model"
	"github.com/rancher/rdns-server/service"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	flags = map[string]map[string]string{
		"DOMAIN":           {"used to set etcd root domain.": "lb.rancher.cloud"},
		"ETCD_ENDPOINTS":   {"used to set etcd endpoints.": "http://127.0.0.1:2379"},
		"ETCD_PREFIX_PATH": {"used to set etcd prefix path.": "/rdnsv3"},
		"ETCD_LEASE_TIME":  {"used to set etcd lease time.": "240h"},
		"CORE_DNS_FILE":    {"used to set coredns file.": "/etc/rdns/config/Corefile"},
		"CORE_DNS_PORT":    {"used to set coredns port.": "53"},
		"CORE_DNS_CPU":     {"used to set coredns cpu, a number (e.g. 3) or a percent (e.g. 50%).": "50%"},
		"CORE_DNS_DB_FILE": {"used to set coredns file plugin db's file name (e.g. /etc/rdns/config/dbfile).": ""},
		"CORE_DNS_DB_ZONE": {"used to set coredns file plugin db's zone (e.g. api.lb.rancher.cloud).": ""},
		"TTL":              {"used to set coredns ttl.": "60"},
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

	if err := generateCoreFile(); err != nil {
		return err
	}

	done := make(chan struct{})

	go metric.StartMetricDaemon(done)

	go coredns.StartCoreDNSDaemon()

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
			if k == "CORE_DNS_DB_FILE" || k == "CORE_DNS_DB_ZONE" {
				continue
			}
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

func generateCoreFile() error {
	fp := os.Getenv("CORE_DNS_FILE")
	if fp == "" {
		return errors.New("failed to get core dns file")
	}
	_, err := os.Stat(fp)
	if err != nil {
		// render CoreFile template
		cf := &model.CoreFile{
			CoreDNSDBFile:  os.Getenv("CORE_DNS_DB_FILE"),
			CoreDNSDBZone:  os.Getenv("CORE_DNS_DB_ZONE"),
			Domain:         os.Getenv("DOMAIN"),
			EtcdPrefixPath: os.Getenv("ETCD_PREFIX_PATH"),
			EtcdEndpoints:  strings.Join(strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","), " "),
			TTL:            os.Getenv("TTL"),
			WildCardBound:  strconv.Itoa(len(strings.Split(strings.TrimRight(os.Getenv("DOMAIN"), "."), ".")) + 1),
		}
		p := template.Must(template.New("corefile-tmpl").Parse(model.CoreFileTmpl))
		f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := p.Execute(f, cf); err != nil {
			return err
		}
	}
	return nil
}
