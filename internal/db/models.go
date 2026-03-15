package db

import (
	"database/sql"
)

type DbHandler struct {
	Db     *sql.DB
	DbName string
}
