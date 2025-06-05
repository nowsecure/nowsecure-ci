package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunFileCommand(v *viper.Viper) *cobra.Command {
	var fileCmd = &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := zerolog.Ctx(ctx)
			fileName := args[0]
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}

			config, err := internal.NewRunConfig(v)
			if err != nil {
				return err
			}

			ctx = internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())

			client, err := internal.ClientFromConfig(config, nil)
			if err != nil {
				return err
			}

			w, err := output.New(config.Output, config.OutputFormat)
			if err != nil {
				zerolog.Ctx(ctx).Panic().Err(err).Msg("Failed to create writer")
			}
			defer w.Close()

			buildResponse, err := uploadFile(ctx, file, config, client)
			if err != nil {
				return err
			}

			if config.PollForMinutes <= 0 {
				log.Info().Msg("Succeeded")
				err = w.Write(buildResponse)
				return err
			}

			taskResponse, err := pollForResults(ctx, client, buildResponse.Package, buildResponse.Platform, buildResponse.Task, config.PollForMinutes)
			if err != nil {
				return err
			}

			if !isAboveMinimum(taskResponse, config.MinimumScore) {
				return fmt.Errorf("the score %.2f is less than the required minimum %d", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
			}

			log.Info().Interface("Assessment", taskResponse.JSON2XX).Msg("Succeeded")

			return w.Write(taskResponse.JSON2XX)
		},
	}

	return fileCmd
}

func uploadFile(ctx context.Context, file *os.File, config internal.RunConfig, client *platformapi.ClientWithResponses) (*platformapi.PostBuild2XX1, error) {
	zerolog.Ctx(ctx).Debug().Msg("uploading file")

	response, err := client.PostBuildWithBodyWithResponse(ctx, &platformapi.PostBuildParams{
		AnalysisType: (*platformapi.PostBuildParamsAnalysisType)(&config.AnalysisType),
		Group:        &config.Group,
	}, "application/octet-stream", file)
	if err != nil {
		return nil, err
	}

	zerolog.Ctx(ctx).Debug().Int("status", response.StatusCode()).Msg("Received http response")

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, response.JSON4XX
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, response.JSON5XX
	}

	buildResponse := platformapi.PostBuild2XX1{}

	err = json.Unmarshal(response.Body, &buildResponse)

	return &buildResponse, err
}
