package backend

import (
	"github.com/rancher/rdns-server/model"

	"github.com/sirupsen/logrus"
)

var currentBackend Backend

type Backend interface {
	Get(opts *model.DomainOptions) (model.Domain, error)
	Set(opts *model.DomainOptions) (model.Domain, error)
	Update(opts *model.DomainOptions) (model.Domain, error)
	Delete(opts *model.DomainOptions) error
	Renew(opts *model.DomainOptions) (model.Domain, error)
	SetText(opts *model.DomainOptions) (model.Domain, error)
	GetText(opts *model.DomainOptions) (model.Domain, error)
	UpdateText(opts *model.DomainOptions) (model.Domain, error)
	DeleteText(opts *model.DomainOptions) error
	SetCNAME(opts *model.DomainOptions) (model.Domain, error)
	GetCNAME(opts *model.DomainOptions) (model.Domain, error)
	UpdateCNAME(opts *model.DomainOptions) (model.Domain, error)
	DeleteCNAME(opts *model.DomainOptions) error
	GetToken(fqdn string) (string, error)
	GetTokenCount() (int64, error)
	GetZone() string
	GetName() string
	MigrateFrozen(opts *model.MigrateFrozen) error
	MigrateToken(opts *model.MigrateToken) error
	MigrateRecord(opts *model.MigrateRecord) error
}

func SetBackend(b Backend) {
	currentBackend = b
}

func GetBackend() Backend {
	if currentBackend == nil {
		logrus.Fatal("not found any backend")
	}
	return currentBackend
}
