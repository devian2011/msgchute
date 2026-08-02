package repository

import (
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// setupMockDB initializes a mock database connection with driver.Valuer support.
// Since it lives in repository_test.go, it is automatically visible to all *_test.go
// files within the repository package.
func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(
		sqlmock.ValueConverterOption(driver.DefaultParameterConverter),
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	require.NoError(t, err)

	db := sqlx.NewDb(mockDB, "postgres")
	return db, mock
}
