package run

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunFileCommand(v *viper.Viper) *cobra.Command {
	// fileCmd represents the file command
	var fileCmd = &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			fileName := args[0]
			file, err := os.Open(fileName)
			if err != nil {
				zerolog.Ctx(ctx).Panic().Err(err).Msgf("Cannot open file %s", fileName)
			}

			config, configErr := internal.NewRunConfig(v)

			if configErr != nil {
				zerolog.Ctx(ctx).Panic().Err(configErr).Msg("Error creating config")
			}

			ctx = zerolog.New(internal.ConsoleLevelWriter{}).
				With().
				Timestamp().
				Logger().
				Level(config.LogLevel).
				WithContext(cmd.Context())

			client, clientErr := internal.ClientFromConfig(config, nil)

			if clientErr != nil {
				zerolog.Ctx(ctx).Panic().Err(clientErr).Msg("Error creating NowSecure API client")
			}

			buildResponse, buildErr := submitFile(ctx, file, config, client)

			if buildErr != nil {
				zerolog.Ctx(ctx).Panic().Err(buildErr).Msg("Error submitting file for assessment")
			}

			if config.PollForMinutes <= 0 {
				// TODO this should probably pretty-print the build response instead of relying on structured logs
				zerolog.Ctx(ctx).Info().Interface("Build Response", buildResponse).Msg("Succeeded")
				return
			}

			taskResponse, taskErr := pollForResults(ctx, client, buildResponse.Package, buildResponse.Platform, buildResponse.Task, config.PollForMinutes)

			if taskErr != nil {
				zerolog.Ctx(ctx).Panic().Err(taskErr).Msg("Error while polling for assessment results")
			}

			if !isAboveMinimum(taskResponse, config.MinimumScore) {
				zerolog.Ctx(ctx).Panic().Msgf("The score %.2f is less than the required minimum %d", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
			}

			// TODO this should probably pretty-print the build response instead of relying on structured logs
			zerolog.Ctx(ctx).Info().Interface("Assessment", taskResponse).Msg("Succeeded")
		},
	}

	return fileCmd
}

func submitFile(ctx context.Context, file *os.File, config internal.RunConfig, client *platformapi.ClientWithResponses) (*platformapi.PostBuild2XX1, error) {
	zerolog.Ctx(ctx).Debug().Msg("uploading file")

	response, responseError := client.PostBuildWithBodyWithResponse(ctx, &platformapi.PostBuildParams{
		AnalysisType: (*platformapi.PostBuildParamsAnalysisType)(&config.AnalysisType),
		Group:        &config.Group,
	}, "application/octet-stream", file)

	if responseError != nil {
		return nil, responseError
	}

	zerolog.Ctx(ctx).Debug().Int("status", response.StatusCode()).Msg("Received http response")

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, errors.New(*response.JSON4XX.Description)
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, errors.New(*response.JSON5XX.Description)
	}

	buildResponse := platformapi.PostBuild2XX1{}

	err := json.Unmarshal(response.Body, &buildResponse)

	return &buildResponse, err
}
