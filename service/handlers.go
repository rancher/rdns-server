package service

import (
	"net/http"

	"github.com/gorilla/context"
)

func apiHandler(f http.Handler) http.Handler {
	return context.ClearHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
}

func createDomain(w http.ResponseWriter, r *http.Request) {

}

func getDomain(w http.ResponseWriter, r *http.Request) {

}

func renewDomain(w http.ResponseWriter, r *http.Request) {

}

func updateDomain(w http.ResponseWriter, r *http.Request) {

}

func deleteDomain(w http.ResponseWriter, r *http.Request) {

}
