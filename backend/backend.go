package backend

import (
	"github.com/Sirupsen/logrus"
)

var currentBackend *Backend

type Backend interface {
	Get()
	List()
	Create()
	Delete()
	Update()
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
