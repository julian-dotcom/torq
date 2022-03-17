package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PgConnect connects to a pg database
func PgConnect(dbName, user, password, host, port string) (db *sqlx.DB, err error) {

	db, err = sqlx.Connect("postgres",
		fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port,
			dbName))
	//fmt.Sprintf("user=%s dbname=%s password=%s port=%s host=%s sslmode=disable", user, dbName,
	//	password, host, port))
	if err != nil {
		return nil, fmt.Errorf("internal/database/connect PgConnect: %v", err)
	}
	return db, nil
}
