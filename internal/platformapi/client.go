package platformapi

import (
	"context"
	"net/http"
)

// TODO pass in config for host and useragent
func New(doer HttpRequestDoer) (*ClientWithResponses, error) {
	if doer == nil {
		doer = &http.Client{}
	}

	host := ""
	userAgent := ""

	return NewClientWithResponses(host,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", userAgent)
			return nil
		}), WithHTTPClient(doer))
}
