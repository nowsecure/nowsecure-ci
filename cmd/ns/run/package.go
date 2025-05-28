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
	// packageCmd represents the package command
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, _ := internal.NewRunConfig(v)

			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())

			packageName := args[0]

			client, err := internal.ClientFromConfig(config, nil)
			zerolog.Ctx(ctx).Debug().Msg("Client created")

			if err != nil {
				zerolog.Ctx(ctx).Panic().Err(err).Msg("Error creating NowSecure API client")
			}

			response, assessmentErr := triggerAssessment(ctx, packageName, config, client)

			if assessmentErr != nil {
				zerolog.Ctx(ctx).Panic().Err(err).Msg("Error triggering assessment")
			}

			if config.PollForMinutes <= 0 {
				zerolog.Ctx(ctx).Info().Any("Assessment response", response).Msg("Succeeded")
				return
			}

			taskResponse, taskErr := pollForResults(ctx, client, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task), config.PollForMinutes)

			if taskErr != nil {
				zerolog.Ctx(ctx).Panic().Err(taskErr).Msg("Error while polling for assessment results")
			}

			isAboveMinimum(taskResponse, config.MinimumScore)

			// TODO this should probably pretty-print the build response instead of relying on structured logs
			zerolog.Ctx(ctx).Info().Any("Assessment", taskResponse).Msg("Succeeded")
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
	zerolog.Ctx(ctx).Debug().Str("pakcage", packageName).Str("platform", config.Platform).Msg("Triggering assessment")
	response, responseError := client.PostAppPlatformPackageAssessmentWithResponse(
		ctx,
		platformapi.PostAppPlatformPackageAssessmentParamsPlatform(config.Platform),
		packageName,
		&platformapi.PostAppPlatformPackageAssessmentParams{
			AnalysisType: (*platformapi.PostAppPlatformPackageAssessmentParamsAnalysisType)(&config.AnalysisType),
		},
	)

	if responseError != nil {
		return nil, responseError
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		zerolog.Ctx(ctx).Debug().Any("Response", response.JSON4XX).Msg("Bad Request")
		return nil, errors.New(*response.JSON4XX.Message)
	}

	if response.HTTPResponse.StatusCode >= 500 {
		zerolog.Ctx(ctx).Debug().Any("Response", response.JSON5XX).Msg("Server error")
		return nil, errors.New(*response.JSON5XX.Message)
	}

	zerolog.Ctx(ctx).Debug().Any("Response", response.JSON2XX).Msg("Successfully triggered assessment")

	return response, nil
}
