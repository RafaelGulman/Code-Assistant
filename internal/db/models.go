package db

import (
	"database/sql"
)

type DbHandler struct {
	DbName string
	Db     *sql.DB
}
