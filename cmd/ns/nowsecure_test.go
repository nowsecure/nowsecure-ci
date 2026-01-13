package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
)

func setupTest(t *testing.T) (*viper.Viper, *internal.BaseConfig, context.Context) {
	v := viper.New()
	config := &internal.BaseConfig{}
	ctx := t.Context()

	return v, config, ctx
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, out string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func TestCommandFromFlags(t *testing.T) {
	groupRef := uuid.New()
	token := "test-token"
	host := "https://localhost:8080"
	ciEnv := "jenkins"
	logLevel := "debug"
	outputPath := "./some-output"
	outputFormat := "json"
	uiHost := "https://localhost:8081"

	t.Run("Long base flags bind correctly", func(t *testing.T) {
		args := []string{
			"--api-host", host,
			"--ci-environment", ciEnv,
			"--token", token,
			"--group-ref", groupRef.String(),
			"--log-level", logLevel,
			"--output", outputPath,
			"--output-format", outputFormat,
			"--ui-host", uiHost,
		}

		v, config, ctx := setupTest(t)
		rootCmd := RootCommand(ctx, v, config)

		_, _, err := executeCommandC(rootCmd, append(args, "help")...)

		require.NoError(t, err)

		assert.Equal(t, host, v.GetString("api_host"))
		assert.Equal(t, ciEnv, v.GetString("ci_environment"))
		assert.Equal(t, token, v.GetString("token"))
		assert.Equal(t, groupRef.String(), v.GetString("group_ref"))
		assert.Equal(t, logLevel, v.GetString("log_level"))
		assert.Equal(t, outputPath, v.GetString("output"))
		assert.Equal(t, outputFormat, v.GetString("output_format"))
		assert.Equal(t, uiHost, v.GetString("ui_host"))

		assert.Equal(t, groupRef.String(), config.Group.String())
		assert.Equal(t, logLevel, config.LogLevel.String())
		assert.Equal(t, output.JSON, config.OutputFormat)
		assert.Equal(t, outputPath, config.Output)
		assert.Equal(t, uiHost, config.UIHost)
	})

	t.Run("Short base flags bind correctly", func(t *testing.T) {
		args := []string{
			"-o", outputPath,
			"-v",
		}

		v, config, ctx := setupTest(t)
		rootCmd := RootCommand(ctx, v, config)
		_, _, err := executeCommandC(rootCmd, args...)

		require.NoError(t, err)

		assert.Equal(t, outputPath, v.GetString("output"))
		assert.True(t, v.GetBool("verbose"))
	})

	t.Run("Log level and verbose cannot both be set", func(t *testing.T) {
		v, config, ctx := setupTest(t)
		rootCmd := RootCommand(ctx, v, config)
		_, _, err := executeCommandC(rootCmd, "--token", "some-token", "--verbose", "--log-level", "debug", "help")
		require.ErrorContains(t, err, "if any flags in the group [log-level verbose] are set none of the others can be")
	})

	t.Run("Invalid config path throws error", func(t *testing.T) {
		v, config, ctx := setupTest(t)
		rootCmd := RootCommand(ctx, v, config)
		_, _, err := executeCommandC(rootCmd, "--config", "./some/bad/path", "help")
		require.ErrorContains(t, err, "no such file or directory")
	})
}

func TestCommandFromEnvVars(t *testing.T) {
	groupRef := uuid.New()
	token := "test-token"
	host := "https://localhost:8080"
	ciEnv := "jenkins"
	logLevel := "debug"
	outputPath := "./some-output"
	outputFormat := "json"
	uiHost := "https://localhost:8081"

	t.Run("Base envvars bind correctly", func(t *testing.T) {
		v, config, ctx := setupTest(t)

		envVars := map[string]string{
			"NS_API_HOST":       host,
			"NS_CI_ENVIRONMENT": ciEnv,
			"NS_GROUP_REF":      groupRef.String(),
			"NS_LOG_LEVEL":      logLevel,
			"NS_OUTPUT":         outputPath,
			"NS_OUTPUT_FORMAT":  outputFormat,
			"NS_TOKEN":          token,
			"NS_UI_HOST":        uiHost,
		}

		for key, value := range envVars {
			os.Setenv(key, value)
			//nolint:gocritic // unsetting intentionally at function end instead of iteration
			defer os.Unsetenv(key)
		}

		rootCmd := RootCommand(ctx, v, config)
		_, _, err := executeCommandC(rootCmd, "help")

		require.NoError(t, err)

		assert.Equal(t, host, v.GetString("api_host"))
		assert.Equal(t, ciEnv, v.GetString("ci_environment"))
		assert.Equal(t, token, v.GetString("token"))
		assert.Equal(t, groupRef.String(), v.GetString("group_ref"))
		assert.Equal(t, logLevel, v.GetString("log_level"))
		assert.Equal(t, outputPath, v.GetString("output"))
		assert.Equal(t, outputFormat, v.GetString("output_format"))
		assert.Equal(t, uiHost, v.GetString("ui_host"))

		assert.Equal(t, uiHost, config.UIHost)
		assert.Equal(t, groupRef.String(), config.Group.String())
		assert.Equal(t, logLevel, config.LogLevel.String())
	})

	t.Run("Invalid config path throws error", func(t *testing.T) {
		v, config, ctx := setupTest(t)

		envVars := map[string]string{
			"NS_CONFIG": "./some/bad/path",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
			//nolint:gocritic // unsetting intentionally at function end instead of iteration
			defer os.Unsetenv(key)
		}

		rootCmd := RootCommand(ctx, v, config)
		_, _, err := executeCommandC(rootCmd, "help")
		require.ErrorContains(t, err, "no such file or directory")
	})
}

func TestCommandFromConfig(t *testing.T) {
	groupRef := uuid.New()
	token := "test-token"
	host := "https://localhost:8080"
	ciEnv := "jenkins"
	logLevel := "debug"
	outputPath := "./some-output"
	outputFormat := "json"
	uiHost := "https://localhost:8081"

	t.Run("Base config file bind correctly", func(t *testing.T) {
		configContent := map[string]string{
			"token":          token,
			"api_host":       host,
			"ui_host":        uiHost,
			"group_ref":      groupRef.String(),
			"log_level":      logLevel,
			"output":         outputPath,
			"output_format":  outputFormat,
			"ci_environment": ciEnv,
		}

		data, err := yaml.Marshal(configContent)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, ".ns-ci.yaml")
		err = os.WriteFile(configFile, data, 0o600)
		require.NoError(t, err)

		v, config, ctx := setupTest(t)

		rootCmd := RootCommand(ctx, v, config)
		_, _, err = executeCommandC(rootCmd, "--config", configFile, "help")

		require.NoError(t, err)

		assert.Equal(t, host, v.GetString("api_host"))
		assert.Equal(t, ciEnv, v.GetString("ci_environment"))
		assert.Equal(t, token, v.GetString("token"))
		assert.Equal(t, groupRef.String(), v.GetString("group_ref"))
		assert.Equal(t, logLevel, v.GetString("log_level"))
		assert.Equal(t, outputPath, v.GetString("output"))
		assert.Equal(t, outputFormat, v.GetString("output_format"))
		assert.Equal(t, uiHost, v.GetString("ui_host"))

		assert.Equal(t, uiHost, config.UIHost)
		assert.Equal(t, groupRef.String(), config.Group.String())
		assert.Equal(t, logLevel, config.LogLevel.String())
	})
}
