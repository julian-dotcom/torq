package database

var (
	SqlUniqueConstraintError    string = "Unique constraint violation."                //nolint:gochecknoglobals
	SqlExecutionError           string = "Executing SQL statement."                    //nolint:gochecknoglobals
	SqlAffectedRowsCheckError   string = "Checking the affected rows."                 //nolint:gochecknoglobals
	SqlBeginTransactionError    string = "Starting SQL transaction."                   //nolint:gochecknoglobals
	SqlRollbackTransactionError string = "Rollback SQL transaction."                   //nolint:gochecknoglobals
	SqlCommitTransactionError   string = "Committing SQL transaction."                 //nolint:gochecknoglobals
	SqlScanResulSetError        string = "Reading the results from the SQL statement." //nolint:gochecknoglobals
)
