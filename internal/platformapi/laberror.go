package platformapi

import (
	"errors"
	"fmt"
	"strconv"

	nserrors "github.com/nowsecure/nowsecure-ci/internal/errors"
)

var _ nserrors.CIError = (*LabRouteError)(nil)

func (w *LabRouteError) Error() string {
	return fmt.Sprintf("HTTP %s - %s: %s", *w.Status, *w.Name, *w.Message)
}

func (w *LabRouteError) ExitCode() int {
	return 1
}

func (w *LabRouteError) StatusCode() (int, error) {
	if w.Status != nil {
		i, err := strconv.Atoi(*w.Status)
		if err != nil {
			return i, nil
		}
		return 0, err
	}

	return 0, errors.New("status code not defined on response")
}
