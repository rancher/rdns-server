package database

import (
	"time"

	"github.com/rancher/rdns-server/model"

	"github.com/sirupsen/logrus"
)

var currentDatabase Database

type Database interface {
	InsertFrozen(prefix string) error
	QueryFrozen(prefix string) (string, error)
	RenewFrozen(prefix string) error
	DeleteFrozen(prefix string) error
	DeleteExpiredFrozen(*time.Time) error
	MigrateFrozen(prefix string, expiration int64) error
	InsertToken(token, name string) (int64, error)
	QueryTokenCount() (int64, error)
	QueryToken(name string) (*model.Token, error)
	QueryExpiredTokens(*time.Time) ([]*model.Token, error)
	RenewToken(name string) (int64, int64, error)
	DeleteToken(prefix string) error
	MigrateToken(token, name string, expiration int64) error
	InsertA(*model.RecordA) (int64, error)
	UpdateA(*model.RecordA) (int64, error)
	QueryA(name string) (*model.RecordA, error)
	ListSubA(id int64) ([]*model.SubRecordA, error)
	DeleteA(name string) error
	InsertSubA(*model.SubRecordA) (int64, error)
	UpdateSubA(*model.SubRecordA) (int64, error)
	QuerySubA(name string) (*model.SubRecordA, error)
	DeleteSubA(name string) error
	InsertCNAME(*model.RecordCNAME) (int64, error)
	UpdateCNAME(*model.RecordCNAME) (int64, error)
	QueryCNAME(name string) (*model.RecordCNAME, error)
	DeleteCNAME(name string) error
	InsertTXT(*model.RecordTXT) (int64, error)
	UpdateTXT(*model.RecordTXT) (int64, error)
	QueryTXT(name string) (*model.RecordTXT, error)
	QueryExpiredTXTs(id int64) ([]*model.RecordTXT, error)
	DeleteTXT(name string) error
	Close() error
}

func SetDatabase(d Database) {
	currentDatabase = d
}

func GetDatabase() Database {
	if currentDatabase == nil {
		logrus.Fatal("not found any database")
	}
	return currentDatabase
}
