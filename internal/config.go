package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/viper"

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

type BaseConfig struct {
	Host      string
	Token     string
	Group     string
	UserAgent string
}

func NewConfig(v *viper.Viper) BaseConfig {
	host := v.GetString("host")
	token := v.GetString("token")

	if host == "" || token == "" {
		fmt.Println("Host and token must both be specified either in a config file, or through a flag")
		os.Exit(1)
	}

	return BaseConfig{
		Host:      host,
		Token:     token,
		Group:     v.GetString("group"),
		UserAgent: v.GetString("userAgent"),
	}
}
