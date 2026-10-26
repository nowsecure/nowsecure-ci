package internal

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/version"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

type BaseConfig struct {
	APIHost        string
	UIHost         string
	PlatformClient platformapi.ClientWithResponsesInterface
	Group          uuid.UUID
	LogLevel       zerolog.Level
	Output         string
	OutputFormat   output.Formats
	UserAgent      string
}

type RunConfig struct {
	BaseConfig
	AnalysisType         string
	PollForMinutes       int
	PollingInterval      time.Duration
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

	APIHost, err := validateHost("api_host", APIHost)
	if err != nil {
		return nil, err
	}

	UIHost := v.GetString("ui_host")
	if UIHost != "" {
		UIHost, err = validateHost("ui_host", UIHost)
		if err != nil {
			return nil, err
		}
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
	userAgent := strings.TrimSpace(fmt.Sprintf("nowsecure-ci/%s (%s) %s", version.Version(), platformInfo, v.GetString("ci_environment")))

	platformClient, err := platformapi.ClientFromConfig(platformapi.Config{
		Host:      APIHost,
		UserAgent: userAgent,
		Token:     token,
	}, nil)
	if err != nil {
		return nil, err
	}

	return &BaseConfig{
		UIHost:         UIHost,
		PlatformClient: platformClient,
		Group:          group,
		LogLevel:       logLevel,
		Output:         v.GetString("output"),
		OutputFormat:   format,
		UserAgent:      userAgent,
	}, nil
}

var allowedHostDomains = []string{"nowsecure.com", "nowsecure.io"}

// validateHost ensures a host URL is safe to send the API token to: https (or
// http on loopback only), no userinfo, and a NowSecure or loopback host.
func validateHost(name, raw string) (string, error) {
	raw = strings.TrimRight(raw, "/")

	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("invalid %s %q: must be an absolute URL such as https://api.nowsecure.com", name, raw)
	}

	if u.User != nil {
		return "", fmt.Errorf("invalid %s: must not contain user credentials", name)
	}

	host := strings.ToLower(u.Hostname())
	loopback := host == "localhost"
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		loopback = true
	}

	switch u.Scheme {
	case "https":
	case "http":
		if !loopback {
			return "", fmt.Errorf("invalid %s %q: http is only allowed for localhost, use https", name, raw)
		}
	default:
		return "", fmt.Errorf("invalid %s %q: scheme must be https", name, raw)
	}

	if loopback {
		return raw, nil
	}

	for _, domain := range allowedHostDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return raw, nil
		}
	}

	return "", fmt.Errorf("invalid %s %q: host must be a %s domain", name, raw, strings.Join(allowedHostDomains, " or "))
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
		PollingInterval:      time.Minute,
		MinimumScore:         v.GetInt("minimum_score"),
		Platform:             platform,
	}, nil
}
