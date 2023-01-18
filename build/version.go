package build

import (
	_ "embed"
)

//go:generate bash get_version.sh
//go:embed version.txt
var version string //nolint:gochecknoglobals

func ExtendedVersion() string {
	return version
}
