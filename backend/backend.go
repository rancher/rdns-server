package backend

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

var currentBackend *Backend

type Backend interface {
	Get(dopts *DomainOptions) (Domain, error)
	Create(dopts *DomainOptions) (Domain, error)
	Update(dopts *DomainOptions) (Domain, error)
	Delete(dopts *DomainOptions) error
	Renew(dopts *DomainOptions) (Domain, error)
}

type Domain struct {
	Fqdn       string     `"json:fqdn"`
	Hosts      []string   `"json:hosts"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

type DomainOptions struct {
	Fqdn  string   `"json:fqdn"`
	Hosts []string `"json:hosts"`
}

func ParseDomainOptions(r *http.Request) (opts *DomainOptions, err error) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(opts)
	return opts, err
}

func SetBackend(b *Backend) {
	currentBackend = b
}

func GetBackend() *Backend {
	if currentBackend == nil {
		logrus.Fatal("Not found any backend")
	}
	return currentBackend
}
