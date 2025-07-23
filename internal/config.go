package internal

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/output"
)

type BaseConfig struct {
	APIHost      string
	UIHost       string
	Token        string
	Group        uuid.UUID
	UserAgent    string
	LogLevel     zerolog.Level
	Output       string
	OutputFormat output.Formats
	ArtifactsDir string
}

type RunConfig struct {
	BaseConfig
	AnalysisType   string
	PollForMinutes int
	MinimumScore   int
	Platform       string
	WithFindings   bool
}

type GetConfig struct {
	BaseConfig
}

func NewBaseConfig(v *viper.Viper) (*BaseConfig, error) {
	APIHost := v.GetString("api_host")
	token := v.GetString("token")

	if APIHost == "" {
		return nil, errors.New("API host must be specified either in a config file, the api_host envvar, or through the --api-host flag")
	}

	if token == "" {
		return nil, errors.New("token must be specified either in a config file, an envvar, or through a flag")
	}

	logLevel, err := zerolog.ParseLevel(v.GetString("log_level"))
	if err != nil {
		return nil, err
	}

	if v.GetBool("verbose") {
		logLevel = zerolog.DebugLevel
	}

	group := uuid.Nil
	if v.IsSet("group_ref") {
		var err error
		group, err = uuid.Parse(v.GetString("group_ref"))
		if err != nil {
			return nil, fmt.Errorf("invalid group_ref: %w", err)
		}
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

	platformInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	userAgent := strings.TrimSpace(fmt.Sprintf("nowsecure-ci/%s (%s) %s", "0.2.0", platformInfo, v.GetString("ci_environment")))

	return &BaseConfig{
		APIHost:      APIHost,
		UIHost:       v.GetString("ui_host"),
		Token:        token,
		Group:        group,
		UserAgent:    userAgent,
		ArtifactsDir: v.GetString("artifacts_dir"),
		LogLevel:     logLevel,
		Output:       v.GetString("output"),
		OutputFormat: format,
	}, nil

}

func NewRunConfig(v *viper.Viper) (*RunConfig, error) {

	baseConfig, err := NewBaseConfig(v)
	if err != nil {
		return nil, err
	}

	if v.IsSet("fetch_artifacts") && v.GetInt("poll_for_minutes") <= 0 {
		return nil, fmt.Errorf("cannot set fetch_artifacts without setting a nonzero poll_for_minutes")
	}

	platform := ""

	if v.IsSet("platform_android") {
		platform = "android"
	}

	if v.IsSet("platform_ios") {
		platform = "ios"
	}

	return &RunConfig{
		BaseConfig:     *baseConfig,
		AnalysisType:   v.GetString("analysis_type"),
		WithFindings:   v.GetBool("with_findings"),
		PollForMinutes: v.GetInt("poll_for_minutes"),
		MinimumScore:   v.GetInt("minimum_score"),
		Platform:       platform,
	}, nil
}

func NewGetConfig(v *viper.Viper) (*GetConfig, error) {
	baseConfig, err := NewBaseConfig(v)
	if err != nil {
		return nil, err
	}

	return &GetConfig{
		BaseConfig: *baseConfig,
	}, nil
}
