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
		Level(zerolog.FatalLevel).
		WithContext(context.Background())

	configureFlags(ctx)

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Panic().Err(err)
	}
}

func configureFlags(ctx context.Context) {
	configFile := ""
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.ns-ci)")

	v := initViper(ctx, configFile)

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
		zerolog.Ctx(ctx).Panic().Err(errs).Msg("Failed binding root level flags")
	}

	rootCmd.AddCommand(run.NewRunCommand(ctx, v))
}

func initViper(ctx context.Context, configFile string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("NS")
	v.AutomaticEnv()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defaultName := ".ns-ci"

		configPath := filepath.Join(home, defaultName)
		if _, err := os.Stat(configPath); err != nil {
			zerolog.Ctx(ctx).Info().Msgf("Config file '%s' does not exist. Continuing with default config", configPath)
			return v
		}

		v.AddConfigPath(home)
		v.SetConfigName(defaultName)
		v.SetConfigType("yaml")
	}

	zerolog.Ctx(ctx).Debug().Msgf("Using config file '%s'", v.ConfigFileUsed())

	err := v.ReadInConfig()
	if err != nil {
		zerolog.Ctx(ctx).Panic().Err(err).Msg("Failed to read in config file %s")
	}

	return v
}
