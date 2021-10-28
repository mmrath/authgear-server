// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package plan

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// Injectors from wire.go:

func NewService(ctx context.Context, pool *db.Pool, databaseCredentials *config.DatabaseCredentials) *Service {
	databaseConfig := NewDatabaseConfig()
	databaseEnvironmentConfig := NewDatabaseEnvironmentConfig(databaseCredentials, databaseConfig)
	factory := NewLoggerFactory()
	handle := globaldb.NewHandle(ctx, pool, databaseEnvironmentConfig, factory)
	clock := _wireSystemClockValue
	sqlBuilder := globaldb.NewSQLBuilder(databaseEnvironmentConfig)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	store := &plan.Store{
		Clock:       clock,
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	configsourceStore := &configsource.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	service := &Service{
		Handle:            handle,
		Store:             store,
		ConfigSourceStore: configsourceStore,
		Clock:             clock,
	}
	return service
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
)
