package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// SetupContextFunc setup context for usage in handler
type SetupContextFunc func(context.Context) context.Context

// CleanupContextFunc cleanup context after handler completed
type CleanupContextFunc func(context.Context)

// Server embeds a net/http server and has a gorillax mux internally
type Server struct {
	*http.Server

	router                     *mux.Router
	authContextResolverFactory authn.AuthContextResolverFactory
	dbPool                     db.Pool
	setupCtxFn                 SetupContextFunc
	cleanupCtxFn               CleanupContextFunc
}

// NewServer create a new Server with default option
func NewServer(
	addr string,
	authContextResolverFactory authn.AuthContextResolverFactory,
	dbPool db.Pool,
	setupCtxFn SetupContextFunc,
	cleanupCtxFn CleanupContextFunc,
) Server {
	return NewServerWithOption(
		addr,
		authContextResolverFactory,
		dbPool,
		setupCtxFn,
		cleanupCtxFn,
		DefaultOption(),
	)
}

// NewServerWithOption create a new Server
func NewServerWithOption(
	addr string,
	authContextResolverFactory authn.AuthContextResolverFactory,
	dbPool db.Pool,
	setupCtxFn SetupContextFunc,
	cleanupCtxFn CleanupContextFunc,
	option Option,
) Server {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthCheckHandler)

	var subRouter *mux.Router
	if option.GearPathPrefix == "" {
		subRouter = router.NewRoute().Subrouter()
	} else {
		subRouter = router.PathPrefix(option.GearPathPrefix).Subrouter()
	}
	srv := Server{
		router: subRouter,
		Server: &http.Server{
			Addr:         addr,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      router,
		},
		authContextResolverFactory: authContextResolverFactory,
		dbPool:                     dbPool,
		setupCtxFn:                 setupCtxFn,
		cleanupCtxFn:               cleanupCtxFn,
	}

	if option.RecoverPanic {
		srv.Use(middleware.RecoverMiddleware{
			RecoverHandler: option.RecoverPanicHandler,
		}.Handle)
	}

	return srv
}

// Handle delegates gorilla mux Handler, and accept a HandlerFactory instead of Handler
func (s *Server) Handle(path string, hf handler.Factory) *mux.Route {
	return s.router.NewRoute().Path(path).Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		configuration := config.GetTenantConfig(r)

		// TODO: improve the logger
		log := logging.CreateLoggerWithContext(r.Context(), "server")

		r = auth.InitRequestAuthContext(r)
		r = db.InitRequestDBContext(r, s.dbPool)
		if s.setupCtxFn != nil {
			r = r.WithContext(s.setupCtxFn(r.Context()))
		}
		if s.cleanupCtxFn != nil {
			defer s.cleanupCtxFn(r.Context())
		}

		h := hf.NewHandler(r)

		txContext := db.NewTxContextWithContext(r.Context(), configuration)
		resolver := s.authContextResolverFactory.NewResolver(r.Context(), configuration)
		err := db.WithTx(txContext, func() error {
			return resolver.Resolve(r, auth.NewContextSetterWithContext(r.Context()))
		})
		if err != nil {
			// TODO: log
			log.WithError(err).Error("failed to resolve auth")
			s.handleError(rw, err)
			return
		}

		policy := hf.ProvideAuthzPolicy()
		if err = policy.IsAllowed(r, auth.NewContextGetterWithContext(r.Context())); err != nil {
			// TODO: log
			log.WithError(err).Error("authz not allowed")
			s.handleError(rw, err)
			return
		}

		h.ServeHTTP(rw, r)
	}))
}

// Use set middlewares to underlying router
func (s *Server) Use(mwf ...mux.MiddlewareFunc) {
	s.router.Use(mwf...)
}

func (s *Server) handleError(rw http.ResponseWriter, err error) {
	skyErr := skyerr.MakeError(err)
	httpStatus := nextSkyerr.ErrorDefaultStatusCode(skyErr)
	response := errorResponse{Err: skyErr}
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)
}

// HealthCheckHandler is basic handler for server health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "OK")
}

type errorResponse struct {
	Err skyerr.Error `json:"error,omitempty"`
}
