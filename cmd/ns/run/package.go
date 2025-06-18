package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func PackageCommand(c context.Context, v *viper.Viper) *cobra.Command {
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, _ := internal.NewRunConfig(v)
			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())
			log := zerolog.Ctx(ctx)

			packageName := args[0]

			w, err := output.New(config.Output, config.OutputFormat)
			if err != nil {
				return err
			}
			defer w.Close()

			client, err := platformapi.ClientFromConfig(config, nil)
			if err != nil {
				return err
			}

			response, err := platformapi.TriggerAssessment(ctx, client, platformapi.TriggerAssessmentParams{
				PackageName:  packageName,
				Group:        config.Group,
				AnalysisType: config.AnalysisType,
				Platform:     config.Platform,
			})
			if err != nil {
				return err
			}
			log.Info()Str("URL", fmt.Sprintf("Running assessment with URL: %s/app/%s/assessment/%s", config.UIHost, response.JSON2XX.Application, response.JSON2XX.Ref)).Msg("Assessment URL")

			if config.PollForMinutes <= 0 {
				log.Info().Msg("Succeeded")
				return w.Write(response.JSON2XX)
			}

			ctx, cancel := context.WithTimeout(ctx, time.Duration(config.PollForMinutes)*time.Minute)
			defer cancel()
			taskResponse, err := pollForResults(ctx, client, config.Group, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task))
			if err != nil {
				return err
			}

			if !isAboveMinimum(taskResponse, config.MinimumScore) {
				if err := w.Write(taskResponse.JSON2XX); err != nil {
					return err
				}
				return fmt.Errorf("the score %.2f is less than the required minimum %d", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
			}

			log.Info().Msg("Succeeded")
			return w.Write(taskResponse.JSON2XX)
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
