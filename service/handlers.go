package service

import (
	"encoding/json"
	"net/http"

	"github.com/rancher/rdns-server/backend"
	"github.com/rancher/rdns-server/model"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func returnHTTPError(w http.ResponseWriter, httpStatus int, err error) {
	logrus.Errorf("got a response error: %v", err)
	o := model.Response{
		Status:  httpStatus,
		Message: err.Error(),
	}
	res, _ := json.Marshal(o)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(res)
}

func returnSuccess(w http.ResponseWriter, d model.Domain, msg string) {
	o := model.Response{
		Status:  http.StatusOK,
		Message: msg,
		Data:    d,
	}
	res, err := json.Marshal(o)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func returnSuccessWithToken(w http.ResponseWriter, d model.Domain, msg string) {
	token, err := generateToken(d.Fqdn)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	o := model.Response{
		Status:  http.StatusOK,
		Message: msg,
		Data:    d,
		Token:   token,
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
	o := model.Response{
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
	vals := r.URL.Query()

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	d, err := b.Set(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	returnSuccessWithToken(w, d, "")
}

func getDomain(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]
	msg := ""

	opts := &model.DomainOptions{Fqdn: fqdn}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	d, err := b.Get(opts)
	if err != nil {
		msg = err.Error()
	}
	returnSuccess(w, d, msg)
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

	returnSuccess(w, d, "")
}

func updateDomain(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}
	opts.Fqdn = fqdn

	b := backend.GetBackend()
	d, err := b.Update(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d, "")
}

func deleteDomain(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	err := b.Delete(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}

func createDomainCNAME(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	d, err := b.SetCNAME(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	returnSuccessWithToken(w, d, "")
}

func getDomainCNAME(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]
	msg := ""

	opts := &model.DomainOptions{Fqdn: fqdn}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	d, err := b.GetCNAME(opts)
	if err != nil {
		msg = err.Error()
	}
	returnSuccess(w, d, msg)
}

func updateDomainCNAME(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}
	opts.Fqdn = fqdn

	b := backend.GetBackend()
	d, err := b.UpdateCNAME(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d, "")
}

func deleteDomainCNAME(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	if len(vals["normal"]) > 0 && vals["normal"][0] == "true" {
		opts.Normal = true
	}

	b := backend.GetBackend()
	err := b.DeleteCNAME(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}

func createDomainText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]
	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	opts.Fqdn = fqdn

	b := backend.GetBackend()
	d, err := b.SetText(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d, "")
}

func getDomainText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]
	msg := ""

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	d, err := b.GetText(opts)
	if err != nil {
		msg = err.Error()
	}
	returnSuccess(w, d, msg)
}

func updateDomainText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts, err := model.ParseDomainOptions(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}
	opts.Fqdn = fqdn
	b := backend.GetBackend()
	d, err := b.UpdateText(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccess(w, d, "")
}

func deleteDomainText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fqdn := vars["fqdn"]

	opts := &model.DomainOptions{Fqdn: fqdn}
	b := backend.GetBackend()
	err := b.DeleteText(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}

func ping(w http.ResponseWriter, r *http.Request) {
	returnSuccessNoData(w)
}

func migrateRecord(w http.ResponseWriter, r *http.Request) {
	opts, err := model.ParseMigrateRecord(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	b := backend.GetBackend()
	err = b.MigrateRecord(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}

func migrateFrozen(w http.ResponseWriter, r *http.Request) {
	opts, err := model.ParseMigrateFrozen(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	b := backend.GetBackend()
	err = b.MigrateFrozen(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}

func migrateToken(w http.ResponseWriter, r *http.Request) {
	opts, err := model.ParseMigrateToken(r)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	b := backend.GetBackend()
	err = b.MigrateToken(opts)
	if err != nil {
		returnHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	returnSuccessNoData(w)
}
