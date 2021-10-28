// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// Injectors from wire.go:

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	config := p.ConfigSourceConfig
	factory := p.LoggerFactory
	localFSLogger := configsource.NewLocalFSLogger(factory)
	manager := p.AppBaseResources
	localFS := &configsource.LocalFS{
		Logger:        localFSLogger,
		BaseResources: manager,
		Config:        config,
	}
	databaseLogger := configsource.NewDatabaseLogger(factory)
	environmentConfig := p.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	clock := _wireSystemClockValue
	databaseEnvironmentConfig := &environmentConfig.Database
	sqlBuilder := globaldb.NewSQLBuilder(databaseEnvironmentConfig)
	pool := p.Database
	handle := globaldb.NewHandle(c, pool, databaseEnvironmentConfig, factory)
	sqlExecutor := globaldb.NewSQLExecutor(c, handle)
	store := &configsource.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	database := &configsource.Database{
		Logger:         databaseLogger,
		BaseResources:  manager,
		TrustProxy:     trustProxy,
		Config:         config,
		Clock:          clock,
		Store:          store,
		Database:       handle,
		DatabaseConfig: databaseEnvironmentConfig,
	}
	controller := configsource.NewController(config, localFS, database)
	return controller
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
)
