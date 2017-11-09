package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/niusmallnan/rdns-server/backend"
	"github.com/niusmallnan/rdns-server/model"
)

type HTTPError struct {
	Status  string `"json:status"`
	Message string `"json:msg"`
}

func returnHTTPError(w http.ResponseWriter, httpStatus int, err error) {
	e := HTTPError{
		Status:  strconv.Itoa(httpStatus),
		Message: err.Error(),
	}
	res, _e := json.Marshal(e)
	if _e != nil {
		logrus.Error(_e)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func returnJSON(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	if res != nil {
		w.Write(res)
	}
}

func apiHandler(f http.Handler) http.Handler {
	return context.ClearHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
}

func createDomain(w http.ResponseWriter, r *http.Request) {
	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	b := backend.GetBackend()
	d, err := b.Create(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	res, err := json.Marshal(d)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}
	returnJSON(w, res)
}

func getDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	d, err := b.Get(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	res, err := json.Marshal(d)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}
	returnJSON(w, res)
}

func renewDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	d, err := b.Renew(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	res, err := json.Marshal(d)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}
	returnJSON(w, res)
}

func updateDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}
	opts.Fqdn = fqdn
	b := backend.GetBackend()
	d, err := b.Update(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	res, err := json.Marshal(d)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}
	returnJSON(w, res)
}

func deleteDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	err := b.Delete(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
	}

	returnJSON(w, nil)
}
