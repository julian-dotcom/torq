package build

import (
	_ "embed"
	"strings"
)

//go:generate bash get_version.sh
//go:embed version.txt
var version string //nolint:gochecknoglobals

func ExtendedVersion() string {
	return strings.TrimSpace(version)
}
