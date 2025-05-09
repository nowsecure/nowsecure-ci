package platformapi

import (
	"context"
	"net/http"

	"github.com/nowsecure/nowsecure-ci/internal/util"
)

// TODO pass in config for host and useragent
func New(config util.Config, doer HttpRequestDoer) (*ClientWithResponses, error) {
	if doer == nil {
		doer = &http.Client{}
	}

	return NewClientWithResponses(config.Host,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", config.UserAgent)
			return nil
		}), WithHTTPClient(doer))
}
