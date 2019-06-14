package metric

import (
	"time"

	"github.com/rancher/rdns-server/backend"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	queryDuration = 5 * time.Second

	tokenGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rancher_dns_tokens",
		Help: "The number of the rancher dns tokens",
	})
)

func StartMetricDaemon(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			count, err := backend.GetBackend().GetTokenCount()
			if err != nil {
				logrus.Errorf("failed to count token numbers: %s", err.Error())
			}
			tokenGauge.Set(float64(count))
			time.Sleep(queryDuration)
		}
	}
}
