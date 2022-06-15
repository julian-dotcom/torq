package database

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

func PgConnect(dbName, user, password, host, port string) (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres",
		fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port,
			dbName))

	if err.Error() == "pq: database \""+dbName+"\" does not exist" {
		log.Println("Creating new database")
		return create(dbName, user, password, host, port)
	}
	if err != nil {
		return nil, fmt.Errorf("internal/database/connect PgConnect: %v", err)
	}
	return db, nil
}

func create(dbName, user, password, host, port string) (db *sqlx.DB, err error) {
	default_db, err := sqlx.Connect("postgres",
		fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", "postgres", password, host, port, "postgres"))
	if err != nil {
		return nil, errors.Wrapf(err, "default database connect: ")
	}
	_, err = default_db.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		return nil, errors.Wrapf(err, "database create: ")
	}
	_, err = default_db.Exec("CREATE USER " + user + " WITH ENCRYPTED PASSWORD '" + password + "';")
	if err != nil {
		return nil, errors.Wrapf(err, "database create user: ")
	}
	if _, err = default_db.Exec("GRANT ALL PRIVILEGES ON DATABASE " + dbName + " TO " + user + ";"); err != nil {
		return nil, errors.Wrapf(err, "database create user privileges: ")
	}
	default_db.Close()
	return PgConnect(dbName, user, password, host, port)
}
