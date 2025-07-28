package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
}

type RunConfig struct {
	BaseConfig
	AnalysisType         string
	PollForMinutes       int
	MinimumScore         int
	Platform             string
	FindingsArtifactPath string
	ArtifactsDir         string
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

	if v.IsSet("save_findings") && v.GetInt("poll_for_minutes") <= 0 {
		return nil, fmt.Errorf("cannot set save-findings without setting a nonzero poll-for-minutes")
	}

	platform := ""

	if v.IsSet("platform_android") {
		platform = "android"
	}

	if v.IsSet("platform_ios") {
		platform = "ios"
	}

	artifactsDir := v.GetString("artifacts_dir")
	findingsArtifactPath := ""

	if v.GetBool("save_findings") {
		if err := os.MkdirAll(artifactsDir, os.ModePerm); err != nil {
			return nil, err
		}

		findingsArtifactPath = filepath.Join(artifactsDir, "findings.json")
	}

	return &RunConfig{
		BaseConfig:           *baseConfig,
		AnalysisType:         v.GetString("analysis_type"),
		FindingsArtifactPath: findingsArtifactPath,
		PollForMinutes:       v.GetInt("poll_for_minutes"),
		MinimumScore:         v.GetInt("minimum_score"),
		Platform:             platform,
	}, nil
}
