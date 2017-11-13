package service

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/niusmallnan/rdns-server/backend"
	"github.com/niusmallnan/rdns-server/model"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Status  int          `json:"status"`
	Message string       `json:"msg"`
	Data    model.Domain `json:"data,omitempty"`
}

func returnHTTPError(w http.ResponseWriter, httpStatus int, err error) {
	logrus.Errorf("Got a response error: %v", err)
	o := Response{
		Status:  httpStatus,
		Message: err.Error(),
	}
	res, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(res)
}

func returnSuccess(w http.ResponseWriter, d model.Domain) {
	o := Response{
		Status: http.StatusOK,
		Data:   d,
	}
	res, err := json.Marshal(o)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func returnSuccessNoData(w http.ResponseWriter) {
	o := Response{
		Status: http.StatusOK,
	}
	res, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

}

func apiHandler(f http.Handler) http.Handler {
	return context.ClearHandler(f)
}

func createDomain(w http.ResponseWriter, r *http.Request) {
	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	b := backend.GetBackend()
	d, err := b.Create(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d)
}

func getDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	d, err := b.Get(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d)
}

func renewDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	d, err := b.Renew(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d)
}

func updateDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	opts.Fqdn = fqdn
	b := backend.GetBackend()
	d, err := b.Update(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d)
}

func deleteDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	err := b.Delete(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}
