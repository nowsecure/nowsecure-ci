package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/get"
	"github.com/nowsecure/nowsecure-ci/cmd/ns/run"
	"github.com/nowsecure/nowsecure-ci/cmd/ns/version"
	"github.com/nowsecure/nowsecure-ci/internal"
	nserrors "github.com/nowsecure/nowsecure-ci/internal/errors"
)

var rootCmd = &cobra.Command{
	Use:           "ns",
	Short:         "NowSecure command line tool to interact with NowSecure Platform",
	Version:       version.Version(),
	Long:          ``,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	ctx := zerolog.New(internal.ConsoleLevelWriter{}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.WarnLevel).
		WithContext(context.Background())

	err := configureFlags(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Panic().Err(err).Msg("")
	}

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		if reqErr, ok := err.(nserrors.CIError); ok {
			zerolog.Ctx(ctx).Error().Any("LabRouteError", reqErr).Msg("API Error Response")
			os.Exit(reqErr.ExitCode())
		}
		zerolog.Ctx(ctx).Fatal().Msg(err.Error())
	}
}

func configureFlags(ctx context.Context) error {
	v := viper.New()
	v.SetEnvPrefix("NS")
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	v.SetConfigName(".ns-ci")

	cobra.OnInitialize(func() {
		err := readConfigFile(v)
		if err != nil {
			zerolog.Ctx(ctx).Debug().Err(err).Msg("Error reading from config file")
		}
	})

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().String("api-host", "https://lab-api.nowsecure.com", "REST API base url")
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
		return errs
	}

	rootCmd.MarkFlagsMutuallyExclusive("log-level", "verbose")
	rootCmd.MarkFlagFilename("config")

	rootCmd.AddCommand(run.RunCommand(ctx, v), get.GetCommand(ctx, v))

	return nil
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
	return v.ReadInConfig()
}
