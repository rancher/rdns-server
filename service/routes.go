package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"ping",
		"GET",
		"/ping",
		ping,
	},
	Route{
		"getDomain",
		"GET",
		"/v1/domain/{fqdn}",
		getDomain,
	},
	Route{
		"createDomain",
		"POST",
		"/v1/domain",
		createDomain,
	},
	Route{
		"updateDomain",
		"PUT",
		"/v1/domain/{fqdn}",
		updateDomain,
	},
	Route{
		"deleteDomain",
		"DELETE",
		"/v1/domain/{fqdn}",
		deleteDomain,
	},
	Route{
		"renewDomain",
		"PUT",
		"/v1/domain/{fqdn}/renew",
		renewDomain,
	},
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	logrus.Debugf("Setting HTTP handlers")
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(apiHandler(route.HandlerFunc))
	}

	router.Use(tokenMiddleware)

	return router
}
