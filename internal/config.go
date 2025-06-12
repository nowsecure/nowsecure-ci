package internal

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/output"
)

type BaseConfig struct {
	Host         string
	Token        string
	Group        uuid.UUID
	UserAgent    string
	LogLevel     zerolog.Level
	Output       string
	OutputFormat output.Formats
}

type RunConfig struct {
	BaseConfig
	AnalysisType   string
	PollForMinutes int
	MinimumScore   int
	Platform       string
}

func NewRunConfig(v *viper.Viper) (*RunConfig, error) {
	host := v.GetString("host")
	token := v.GetString("token")

	if host == "" || token == "" {
		return nil, errors.New("host and token must both be specified either in a config file, or through a flag")
	}

	logLevel, err := zerolog.ParseLevel(v.GetString("log_level"))
	if err != nil {
		return nil, err
	}

	if v.GetBool("verbose") {
		logLevel = zerolog.DebugLevel
	}

	group := uuid.Nil
	if v.IsSet("group-ref") {
		var err error
		group, err = uuid.Parse(v.GetString("group"))
		if err != nil {
			return nil, errors.New("must have valid group")
		}
	}

	platform := ""

	if v.IsSet("platform_android") {
		platform = "android"
	}

	if v.IsSet("platform_ios") {
		platform = "ios"
	}

	var format output.Formats
	if v.IsSet("output_format") {
		switch strings.ToLower(v.GetString("output_format")) {
		case "json":
			format = output.JSON
		default:
			return nil, errors.New("must have valid output format")
		}
	}

	return &RunConfig{
		BaseConfig: BaseConfig{
			Host:         host,
			Token:        token,
			Group:        group,
			UserAgent:    v.GetString("user_agent"),
			LogLevel:     logLevel,
			Output:       v.GetString("output"),
			OutputFormat: format,
		},
		AnalysisType:   v.GetString("analysis_type"),
		PollForMinutes: v.GetInt("poll_for_minutes"),
		MinimumScore:   v.GetInt("minimum_score"),
		Platform:       platform,
	}, nil
}
