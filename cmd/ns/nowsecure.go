package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/run"
	"github.com/nowsecure/nowsecure-ci/cmd/ns/version"
	"github.com/nowsecure/nowsecure-ci/internal"
)

func RootCommand(ctx context.Context, v *viper.Viper, config *internal.BaseConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "ns",
		Short:         "NowSecure command line tool to interact with NowSecure Platform",
		Version:       version.Version(),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			v.SetEnvPrefix("NS")
			v.AutomaticEnv()
			v.SetConfigType("yaml")
			v.SetConfigName(".ns-ci")

			if err := readConfigFile(v); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					zerolog.Ctx(ctx).Debug().Msg("No config file found")
				} else {
					zerolog.Ctx(ctx).Debug().Err(err).Msg("Error reading from config file")
					return err
				}
			}

			baseConfig, err := internal.NewBaseConfig(v)
			if err != nil {
				return err
			}
			*config = *baseConfig

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().String("api-host", "https://api.nowsecure.com", "REST API base url")
	rootCmd.PersistentFlags().String("ui-host", "https://app.nowsecure.com", "UI base url")
	rootCmd.PersistentFlags().String("token", "", "auth token for REST API")
	rootCmd.PersistentFlags().String("group-ref", "", "group uuid with which to run assessments")
	rootCmd.PersistentFlags().String("log-level", "info", "logging level")
	rootCmd.PersistentFlags().StringP("output", "o", "", "write  output to <file> instead of stdout.")
	rootCmd.PersistentFlags().String("output-format", "json", "write  output in specified format.")
	rootCmd.PersistentFlags().String("ci-environment", "", "appended to the user_agent header")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging (same as --log-level debug)")
	bindingErrors := []error{
		v.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")),
		v.BindPFlag("api_host", rootCmd.PersistentFlags().Lookup("api-host")),
		v.BindPFlag("ui_host", rootCmd.PersistentFlags().Lookup("ui-host")),
		v.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")),
		v.BindPFlag("group_ref", rootCmd.PersistentFlags().Lookup("group-ref")),
		v.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")),
		v.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output-format")),
		v.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")),
		v.BindPFlag("ci_environment", rootCmd.PersistentFlags().Lookup("ci-environment")),
		v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")),
	}
	if errs := errors.Join(bindingErrors...); errs != nil {
		zerolog.Ctx(ctx).Panic().Err(errs).Msg("Failed binding run level flags")
	}

	rootCmd.MarkFlagsMutuallyExclusive("log-level", "verbose")

	rootCmd.AddCommand(run.RunCommand(ctx, v, config))

	return rootCmd
}

func readConfigFile(v *viper.Viper) error {
	if v.IsSet("config") {
		v.SetConfigFile(v.GetString("config"))
		return v.ReadInConfig()
	}

	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(home)
	}
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return checkWorkingDirConfig(v, home)
}

// workingDirRestrictedKeys may not be set from a config file discovered in the
// working directory, since that directory may be attacker controlled (e.g. a
// checked out repository) and these values determine where the token is sent.
var workingDirRestrictedKeys = []string{"api_host", "ui_host"}

func checkWorkingDirConfig(v *viper.Viper, home string) error {
	configFile := v.ConfigFileUsed()
	if home != "" && sameDir(filepath.Dir(configFile), home) {
		return nil
	}

	for _, key := range workingDirRestrictedKeys {
		if v.InConfig(key) {
			return fmt.Errorf("%s may not be set in a config file discovered in the working directory (%s); use a flag, an NS_ envvar, $HOME/.ns-ci.yaml, or --config instead", key, configFile)
		}
	}

	return nil
}

func sameDir(a, b string) bool {
	resolve := func(p string) string {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			p = resolved
		}
		return p
	}

	return resolve(a) == resolve(b)
}
