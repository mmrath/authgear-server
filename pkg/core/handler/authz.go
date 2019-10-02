package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type authzMiddleware struct {
	authContext    auth.ContextGetter
	policyProvider authz.PolicyProvider
	logger         *logrus.Entry
}

func (m authzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		policy := m.policyProvider.ProvideAuthzPolicy()
		if err := policy.IsAllowed(r, m.authContext); err != nil {
			m.logger.WithError(err).Info("authz not allowed")
			// NOTE(louis): In case the policy returns this error
			// write a header to hint the client SDK to try refresh.
			if err == session.ErrSessionNotFound {
				rw.Header().Set(coreHttp.HeaderTryRefreshToken, "true")
			}
			m.writeUnauthorized(rw, err)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func (m authzMiddleware) writeUnauthorized(rw http.ResponseWriter, err error) {
	skyErr := skyerr.MakeError(err)
	httpStatus := skyerr.ErrorDefaultStatusCode(skyErr)
	response := APIResponse{Err: skyErr}
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)
}

type RequireAuthz func(h http.Handler, p authz.PolicyProvider) http.Handler

func NewRequireAuthzFactory(authCtx auth.ContextGetter, loggerFactory logging.Factory) RequireAuthz {
	return func(h http.Handler, p authz.PolicyProvider) http.Handler {
		m := authzMiddleware{
			authContext:    authCtx,
			policyProvider: p,
			logger:         loggerFactory.NewLogger("authz"),
		}
		return m.Handle(h)
	}
}
