package repository

import (
	"github.com/jmoiron/sqlx"
)

type DBContext interface {
	sqlx.Ext
	sqlx.ExtContext

	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}
