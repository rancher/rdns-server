package backend

import (
	"github.com/niusmallnan/rdns-server/model"
	"github.com/sirupsen/logrus"
)

var currentBackend Backend

type Backend interface {
	Get(dopts *model.DomainOptions) (model.Domain, error)
	Create(dopts *model.DomainOptions) (model.Domain, error)
	Update(dopts *model.DomainOptions) (model.Domain, error)
	Delete(dopts *model.DomainOptions) error
	Renew(dopts *model.DomainOptions) (model.Domain, error)
}

func SetBackend(b Backend) {
	currentBackend = b
}

func GetBackend() Backend {
	if currentBackend == nil {
		logrus.Fatal("Not found any backend")
	}
	return currentBackend
}
