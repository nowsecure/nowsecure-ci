package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/run"
	"github.com/nowsecure/nowsecure-ci/internal"
)

var rootCmd = &cobra.Command{
	Use:   "ns",
	Short: "NowSecure command line tool to interact with NowSecure Platform",
	Long:  ``,
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
		zerolog.Ctx(ctx).Panic().Err(err).Msg("")
	}
}

func configureFlags(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	defaultName := ".ns-ci"

	configPath := filepath.Join(home, defaultName)

	rootCmd.PersistentFlags().StringVar(&configPath, "config", configPath, "config file path")

	v, err := initViper(configPath)

	if err != nil {
		return err
	}

	rootCmd.PersistentFlags().String("host", "https://lab-api.nowsecure.com", "REST API base url")
	rootCmd.PersistentFlags().String("token", "", "auth token for REST API")
	rootCmd.PersistentFlags().StringP("group", "g", "", "group with which to run assessments")
	rootCmd.PersistentFlags().StringP("log-level", "", "info", "logging level")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging (same as --log-level debug)")

	rootCmd.MarkFlagsMutuallyExclusive("log-level", "verbose")

	v.SetDefault("user_agent", "nowsecure-ci")

	bindingErrors := []error{v.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host")),
		v.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")),
		v.BindPFlag("group", rootCmd.PersistentFlags().Lookup("group")),
		v.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")),
		v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")),
	}

	if errs := errors.Join(bindingErrors...); errs != nil {
		return errs
	}

	rootCmd.AddCommand(run.NewRunCommand(ctx, v))

	return nil
}

func initViper(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("NS")
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	err := v.ReadInConfig()

	return v, err
}
