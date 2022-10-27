package database

var (
	SqlUniqueConstraintError    string = "Unique constraint violation."
	SqlExecutionError           string = "Executing SQL statement."
	SqlAffectedRowsCheckError   string = "Checking the affected rows."
	SqlBeginTransactionError    string = "Starting SQL transaction."
	SqlRollbackTransactionError string = "Rollback SQL transaction."
	SqlCommitTransactionError   string = "Committing SQL transaction."
	SqlScanResulSetError        string = "Reading the results from the SQL statement."
)
