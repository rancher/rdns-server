package route53

import (
	"net/http"
	"os"
	"strings"

	"github.com/rancher/rdns-server/backend"
	"github.com/rancher/rdns-server/backend/route53"
	"github.com/rancher/rdns-server/database"
	"github.com/rancher/rdns-server/database/mysql"
	"github.com/rancher/rdns-server/metric"
	"github.com/rancher/rdns-server/purge"
	"github.com/rancher/rdns-server/service"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	flags = map[string]map[string]string{
		"AWS_HOSTED_ZONE_ID":    {"used to set aws hosted zone ID.": ""},
		"AWS_ACCESS_KEY_ID":     {"used to set aws access key ID.": ""},
		"AWS_SECRET_ACCESS_KEY": {"used to set aws secret access key.": ""},
		"DATABASE":              {"used to set database.": "mysql"},
		"DATABASE_LEASE_TIME":   {"used to set database lease time.": "240h"},
		"DSN":                   {"used to set database dsn.": ""},
		"TTL":                   {"used to set route53 ttl.": "10"},
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

	d, err := setDatabase(c)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := setBackend(); err != nil {
		return err
	}

	done := make(chan struct{})

	go metric.StartMetricDaemon(done)

	go purge.StartPurgerDaemon(done)

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

func setDatabase(c *cli.Context) (d *mysql.Database, err error) {
	switch c.String("database") {
	case mysql.DataBaseName:
		d, err = mysql.NewDatabase(c.String("dsn"))
		if err != nil {
			return nil, err
		}
		database.SetDatabase(d)
	default:
		return nil, errors.New("no suitable database found")
	}

	return d, nil
}

func setBackend() error {
	b, err := route53.NewBackend()
	if err != nil {
		return err
	}
	backend.SetBackend(b)

	return nil
}
