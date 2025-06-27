package version

import (
	"strings"
)

var version string

func Version() string {
	return strings.TrimSpace(version)
}
