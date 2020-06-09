// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package auth

import (
	"github.com/gorilla/mux"
	auth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	redis2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	oauth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/provider"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	redis3 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/user"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	pq2 "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"net/http"
)

// Injectors from wire.go:

func NewAccessKeyMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	accessKeyMiddleware := auth.ProvideAccessKeyMiddleware(tenantConfiguration)
	middlewareFunc := provideMiddleware(accessKeyMiddleware)
	return middlewareFunc
}

func NewSessionMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	insecureCookieConfig := ProvideSessionInsecureCookieConfig(m)
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	provider := time.NewProvider()
	factory := logging.ProvideLoggerFactory(context, tenantConfiguration)
	store := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	eventStore := redis2.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, store, accessEventProvider, tenantConfiguration)
	resolver := &session.Resolver{
		CookieConfiguration: cookieConfiguration,
		Provider:            sessionProvider,
		Time:                provider,
	}
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	authorizationStore := &pq.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	resolverSessionProvider := oauth.ProvideResolverProvider(sessionProvider)
	oauthResolver := &oauth.Resolver{
		Authorizations: authorizationStore,
		AccessGrants:   grantStore,
		OfflineGrants:  grantStore,
		Sessions:       resolverSessionProvider,
		Time:           provider,
	}
	authAccessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	middleware := &auth2.Middleware{
		IDPSessionResolver:         resolver,
		AccessTokenSessionResolver: oauthResolver,
		AccessEvents:               authAccessEventProvider,
		AuthInfoStore:              authinfoStore,
		Time:                       provider,
		TxContext:                  txContext,
	}
	middlewareFunc := provideMiddleware(middleware)
	return middlewareFunc
}

func NewCSPMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	middlewareFunc := webapp.ProvideCSPMiddleware(tenantConfiguration)
	return middlewareFunc
}

func NewCSRFMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	middlewareFunc := ProvideCSRFMiddleware(m, tenantConfiguration)
	return middlewareFunc
}

func NewStateMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	context := ProvideContext(r)
	stateStoreImpl := &webapp.StateStoreImpl{
		Context: context,
	}
	middlewareFunc := webapp.ProvideStateMiddleware(stateStoreImpl)
	return middlewareFunc
}

func NewClientIDMiddleware(r *http.Request, m DependencyMap) mux.MiddlewareFunc {
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	middlewareFunc := webapp.ProvideClientIDMiddleware(tenantConfiguration)
	return middlewareFunc
}

func newSessionManager(r *http.Request, m DependencyMap) *auth2.SessionManager {
	context := ProvideContext(r)
	tenantConfiguration := ProvideTenantConfig(context, m)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := &user.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	timeProvider := time.NewProvider()
	reservedNameChecker := ProvideReservedNameChecker(m)
	typeCheckerFactory := loginid.ProvideTypeCheckerFactory(tenantConfiguration, reservedNameChecker)
	checker := loginid.ProvideChecker(tenantConfiguration, typeCheckerFactory)
	normalizerFactory := loginid.ProvideNormalizerFactory(tenantConfiguration)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, timeProvider, tenantConfiguration, checker, normalizerFactory)
	oauthProvider := oauth2.ProvideProvider(sqlBuilder, sqlExecutor, timeProvider)
	anonymousProvider := anonymous.ProvideProvider(sqlBuilder, sqlExecutor)
	providerProvider := provider.ProvideProvider(tenantConfiguration, loginidProvider, oauthProvider, anonymousProvider)
	queries := &user.Queries{
		Store:      store,
		Identities: providerProvider,
		Time:       timeProvider,
	}
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(timeProvider, sqlBuilder, sqlExecutor)
	factory := logging.ProvideLoggerFactory(context, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, tenantConfiguration, txContext, timeProvider, queries, authinfoStore, userprofileStore, loginidProvider, factory)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, timeProvider, factory)
	insecureCookieConfig := ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, timeProvider, tenantConfiguration, cookieConfiguration)
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, timeProvider)
	sessionManager := &oauth.SessionManager{
		Store: grantStore,
		Time:  timeProvider,
	}
	authSessionManager := &auth2.SessionManager{
		Users:               queries,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	return authSessionManager
}

// wire.go:

type middlewareInstance interface {
	Handle(next http.Handler) http.Handler
}

func provideMiddleware(m middlewareInstance) mux.MiddlewareFunc {
	return m.Handle
}
