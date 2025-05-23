package internal

import (
	"context"
	"net/http"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

// TODO pass in config for host and useragent
func ClientFromConfig(doer platformapi.HttpRequestDoer) (*platformapi.ClientWithResponses, error) {
	if doer == nil {
		doer = &http.Client{}
	}

	host := ""
	userAgent := ""

	return platformapi.NewClientWithResponses(host,
		platformapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", userAgent)
			return nil
		}), platformapi.WithHTTPClient(doer))
}
