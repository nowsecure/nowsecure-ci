package run

import (
	"errors"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
)

func NewRunPackageCommand(v *viper.Viper) *cobra.Command {
	// packageCmd represents the package command
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := internal.NewRunConfig(v)

			ctx := zerolog.New(internal.ConsoleLevelWriter{}).
				With().
				Timestamp().
				Logger().
				Level(config.LogLevel).
				WithContext(cmd.Context())

			packageName := args[0]

			zerolog.Ctx(ctx).Panic().Interface("ctx", ctx).Interface("config", config).Str("packageName", packageName).Msg("")
		},
	}

	packageCmd.Flags().Bool("ios", false, "app is for ios platform")
	packageCmd.Flags().Bool("android", false, "app is for android platform")

	packageCmd.MarkFlagsOneRequired("ios", "android")
	packageCmd.MarkFlagsMutuallyExclusive("ios", "android")

	bindingErrors := []error{
		v.BindPFlag("platform_android", packageCmd.Flags().Lookup("android")),
		v.BindPFlag("platform_ios", packageCmd.Flags().Lookup("ios")),
	}

	if errs := errors.Join(bindingErrors...); errs != nil {
		zerolog.Ctx(c).Panic().Err(errs).Msg("Failed binding run level flags")
	}

	return packageCmd
}
