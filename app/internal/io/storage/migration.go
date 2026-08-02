package storage

import (
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
)

func Migrate(cfg *Config) error {
	u, parseErr := url.Parse(cfg.Dsn)
	if parseErr != nil {
		return parseErr
	}
	db := dbmate.New(u)
	db.MigrationsDir = []string{cfg.MigrationDir}

	return db.CreateAndMigrate()
}
