package internal

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func ClientFromConfig(config RunConfig, doer platformapi.HttpRequestDoer) (*platformapi.ClientWithResponses, error) {
	if doer == nil {
		doer = &http.Client{}
	}

	return platformapi.NewClientWithResponses(config.Host,
		platformapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", config.UserAgent)
			req.Header.Add("Authorization", "Bearer "+config.Token)
			return nil
		}), platformapi.WithHTTPClient(doer))
}

type BaseConfig struct {
	Host      string
	Token     string
	Group     uuid.UUID
	UserAgent string
	LogLevel  zerolog.Level
}

type RunConfig struct {
	BaseConfig
	AnalysisType   string
	PollForMinutes int
	MinimumScore   int
	Platform       string
}

func NewRunConfig(v *viper.Viper) (RunConfig, error) {
	host := v.GetString("host")
	token := v.GetString("token")

	if host == "" || token == "" {
		return RunConfig{}, errors.New("host and token must both be specified either in a config file, or through a flag")
	}

	logLevel, err := zerolog.ParseLevel(v.GetString("log_level"))

	if err != nil {
		return RunConfig{}, err
	}

	if v.GetBool("verbose") {
		logLevel = zerolog.DebugLevel
	}

	group := uuid.Nil
	if v.IsSet("group") {
		var err error
		group, err = uuid.Parse(v.GetString("group"))

		if err != nil {
			return RunConfig{}, errors.New("must have valid group")
		}
	}

	platform := ""

	if v.IsSet("platform") {
		platform = strings.ToLower(v.GetString("platform"))

		if platform != "ios" && platform != "android" {
			return RunConfig{}, errors.New("must have valid platform")
		}
	}

	return RunConfig{
		BaseConfig: BaseConfig{
			Host:      host,
			Token:     token,
			Group:     group,
			UserAgent: v.GetString("user_agent"),
			LogLevel:  logLevel,
		},
		AnalysisType:   v.GetString("analysis_type"),
		PollForMinutes: v.GetInt("poll_for_minutes"),
		MinimumScore:   v.GetInt("minimum_score"),
		Platform:       platform,
	}, nil
}
