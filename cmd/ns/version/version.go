package version

import (
	_ "embed"
	"strings"
)

var version string

func Version() string {
	return strings.TrimSpace(version)
}
