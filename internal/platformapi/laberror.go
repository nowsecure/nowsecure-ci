package platformapi

import (
	"fmt"
	nserrors "github.com/nowsecure/nowsecure-ci/internal/errors"
)

var _ nserrors.CIError = (*LabRouteError)(nil)

func (w *LabRouteError) Error() string {
	return fmt.Sprintf("HTTP %s - %s: %s", *w.Status, *w.Name, *w.Message)
}

func (w *LabRouteError) ExitCode() int {
	return 1
}
