package platformapi

import (
	"fmt"
)

func (w *LabRouteError) Error() string {
	return fmt.Sprintf("HTTP %s - %s: %s", *w.Status, *w.Name, *w.Message)
}

func (w *LabRouteError) ExitCode() int {
	return 1
}
