package run

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
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
				log.Info().Any("Assessment response", response.JSON2XX).Msg("Succeeded")
				return nil
			}

			taskResponse, err := pollForResults(ctx, client, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task), config.PollForMinutes)
			if err != nil {
				return err
			}

			isAboveMinimum(taskResponse, config.MinimumScore)

			// TODO this should probably pretty-print the build response instead of relying on structured logs
			log.Info().Any("Assessment", taskResponse).Msg("Succeeded")
			return nil
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
