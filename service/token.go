package service

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/rancher/rdns-server/backend"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// getBaseFqdn can return a base fqdn
// xx.<cluster_id>.lb.rancher.cloud == > <cluster_id>.lb.rancher.cloud
func getBaseFqdn(fqdn string) string {
	slice := strings.Split(fqdn, ".")
	if len(slice) < 4 {
		return fqdn
	}

	return strings.Join(slice[len(slice)-4:], ".")
}

func generateToken(fqdn string) (string, error) {
	b := backend.GetBackend()
	origin, err := b.GetTokenOrigin(fqdn)
	if err != nil {
		logrus.Errorf("Failed to get token origin %s, err: %v", fqdn, err)
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(origin), bcrypt.MinCost)
	if err != nil {
		logrus.Errorf("Failed to generate token with %s, err: %v", fqdn, err)
		return "", err
	}

	token := base64.StdEncoding.EncodeToString(hash)
	return token, nil
}

func compareToken(fqdn, token string) bool {
	// normal text record & acme text record need special treatment
	fqdnLen := len(strings.Split(fqdn, "."))
	rootDomainLen := len(strings.Split(rootDomain, "."))
	diffLen := fqdnLen - rootDomainLen
	if diffLen > 1 {
		sp := strings.SplitAfterN(fqdn, ".", diffLen)
		fqdn = sp[len(sp)-1]
	}

	hash, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		logrus.Errorf("Failed to decode token: %s", fqdn)
		return false
	}

	b := backend.GetBackend()
	origin, err := b.GetTokenOrigin(fqdn)
	if err != nil {
		logrus.Errorf("Failed to get token origin %s, err: %v", fqdn, err)
		return false
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(origin))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"token": token,
			"fqdn":  fqdn,
		}).Errorf("Failed to compare token, err: %v", err)
		return false
	}
	logrus.Debugf("Token **** matched with fqdn %s", fqdn)
	return true
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// createDomain and ping have no need to check token
		logrus.Debugf("Request URL path: %s", r.URL.Path)
		if (r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/txt")) ||
			(r.Method != http.MethodPost && !strings.HasPrefix(r.URL.Path, "/ping")) {
			authorization := r.Header.Get("Authorization")
			token := strings.TrimLeft(authorization, "Bearer ")
			fqdn, ok := mux.Vars(r)["fqdn"]
			if ok {
				if !compareToken(fqdn, token) {
					returnHTTPError(w, http.StatusForbidden, errors.New("Forbidden to use"))
					return
				}
			} else {
				returnHTTPError(w, http.StatusForbidden, errors.New("Must specific the fqdn"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
