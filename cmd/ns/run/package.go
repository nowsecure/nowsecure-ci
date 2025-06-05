package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunPackageCommand(c context.Context, v *viper.Viper) *cobra.Command {
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

			client, err := internal.ClientFromConfig(config, nil)
			if err != nil {
				return err
			}
			log.Debug().Msg("Client created")

			response, err := triggerAssessment(ctx, packageName, config, client)
			if err != nil {
				return err
			}

			if config.PollForMinutes <= 0 {
				log.Info().Msg("Succeeded")
				return w.Write(response.JSON2XX)
			}

			taskResponse, err := pollForResults(ctx, client, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task), config.PollForMinutes)
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

func triggerAssessment(ctx context.Context, packageName string, config internal.RunConfig, client *platformapi.ClientWithResponses) (*platformapi.PostAppPlatformPackageAssessmentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Str("package", packageName).Str("platform", config.Platform).Msg("Triggering assessment")
	response, err := client.PostAppPlatformPackageAssessmentWithResponse(
		ctx,
		platformapi.PostAppPlatformPackageAssessmentParamsPlatform(config.Platform),
		packageName,
		&platformapi.PostAppPlatformPackageAssessmentParams{
			AnalysisType: (*platformapi.PostAppPlatformPackageAssessmentParamsAnalysisType)(&config.AnalysisType),
		},
	)
	if err != nil {
		return nil, err
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		log.Debug().Any("Response", response.JSON4XX).Msg("Bad Request")
		return nil, response.JSON4XX
	}

	if response.HTTPResponse.StatusCode >= 500 {
		log.Debug().Any("Response", response.JSON5XX).Msg("Server error")
		return nil, response.JSON5XX
	}

	log.Debug().Any("Response", response.JSON2XX).Msg("Successfully triggered assessment")

	return response, nil
}
