package database

import "github.com/cockroachdb/errors"

var (
	SqlUniqueConstraintErrorString string = "Unique constraint violation."                         //nolint:gochecknoglobals
	SqlUniqueConstraintError              = errors.New(SqlUniqueConstraintErrorString)             //nolint:gochecknoglobals
	SqlExecutionError              string = "Executing SQL statement."                             //nolint:gochecknoglobals
	SqlAffectedRowsCheckError      string = "Checking the affected rows."                          //nolint:gochecknoglobals
	SqlBeginTransactionError       string = "Starting SQL transaction."                            //nolint:gochecknoglobals
	SqlRollbackTransactionError    string = "Rollback SQL transaction."                            //nolint:gochecknoglobals
	SqlCommitTransactionError      string = "Committing SQL transaction."                          //nolint:gochecknoglobals
	SqlScanResulSetError           string = "Reading the results from the SQL statement."          //nolint:gochecknoglobals
	SqlUpdateOneExecutionError     string = "Update statement updated 0 or more then 1 record(s)." //nolint:gochecknoglobals
)
