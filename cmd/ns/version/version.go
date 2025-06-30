package version

import (
	"strings"
)

var version string

func Version() string {
	if version == "" {
		return "dirty"
	}
	return strings.TrimSpace(version)
}
