package migrations

import "embed"

//go:embed *.psql
var MigrationFiles embed.FS
