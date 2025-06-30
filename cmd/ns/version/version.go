package version

import (
	"strings"
)

var version string

func Version() string {
	if version == "" {
		return "development"
	}
	return strings.TrimSpace(version)
}
