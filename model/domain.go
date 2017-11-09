package model

import (
	"encoding/json"
	"net/http"
	"time"
)

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
	err = decoder.Decode(opts)
	return opts, err
}
