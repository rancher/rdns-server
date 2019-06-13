package purge

import (
	"fmt"
	"os"
	"time"

	"github.com/rancher/rdns-server/backend"
	"github.com/rancher/rdns-server/database"
	"github.com/rancher/rdns-server/model"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	flagFrozen            = "FROZEN"
	flagTTL               = "TTL"
	intervalSeconds int64 = 600
)

type purger struct {
}

func StartPurgerDaemon(done chan struct{}) {
	p := &purger{}
	go wait.JitterUntil(p.purge, time.Duration(intervalSeconds)*time.Second, .1, true, done)
}

func (p *purger) purge() {
	logrus.Debugf("running purge process")

	// check frozen records, delete the frozen record which is expired
	if err := database.GetDatabase().DeleteExpiredFrozen(calculateFrozenTime()); err != nil {
		logrus.Error(err)
	}

	// check token records, delete the token record which is expired
	// this ensures that associated records are also deleted
	tokens, err := database.GetDatabase().QueryExpiredTokens(calculateTTLTime())
	if err != nil {
		logrus.Error(err)
	}

	for _, token := range tokens {
		// delete route53 A records & sub A records & wildcard records
		opts := &model.DomainOptions{
			Fqdn: token.Fqdn,
		}
		a, err := backend.GetBackend().Get(opts)
		if err == nil && a.Fqdn != "" {
			if err := backend.GetBackend().Delete(opts); err != nil {
				logrus.Error(err)
				continue
			}
		}

		// delete route53 TXT records
		ts, err := database.GetDatabase().QueryExpiredTXTs(token.ID)
		for _, t := range ts {
			tOpts := &model.DomainOptions{
				Fqdn: t.Fqdn,
			}
			if err := backend.GetBackend().DeleteText(tOpts); err != nil {
				logrus.Error(err)
				continue
			}
		}

		// delete token records & referenced records
		if err := database.GetDatabase().DeleteToken(token.Token); err != nil {
			logrus.Error(err)
		}
	}
}

func calculateFrozenTime() *time.Time {
	f, err := time.ParseDuration(os.Getenv(flagFrozen))
	if err != nil {
		logrus.Fatalf(errEmptyEnv, flagFrozen)
	}
	d, _ := time.ParseDuration(fmt.Sprintf("%dns", int(f.Nanoseconds())))
	e := time.Now().Add(-d)
	return &e
}

func calculateTTLTime() *time.Time {
	t, err := time.ParseDuration(os.Getenv(flagTTL))
	if err != nil {
		logrus.Fatalf(errEmptyEnv, flagTTL)
	}
	duration, _ := time.ParseDuration(fmt.Sprintf("%dns", int(t.Nanoseconds())))
	e := time.Now().Add(-duration)
	return &e
}
