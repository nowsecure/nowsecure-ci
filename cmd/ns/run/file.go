package run

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func FileCommand(v *viper.Viper, config *internal.BaseConfig) *cobra.Command {
	return &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileName := args[0]
			config, err := internal.NewRunConfig(v)
			if err != nil {
				return err
			}
			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())

			return RunFile(ctx, fileName, config)
		},
	}
}

func RunFile(ctx context.Context, fileName string, config *internal.RunConfig) error {
	log := zerolog.Ctx(ctx)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	client := config.PlatformClient

	w, err := output.New(config.Output, config.OutputFormat)
	if err != nil {
		return err
	}
	defer w.Close()

	buildResponse, err := platformapi.UploadFile(ctx, client, platformapi.UploadFileParams{
		AnalysisType: config.AnalysisType,
		Group:        config.Group,
		File:         file,
	})
	if err != nil {
		return err
	}
	log.Info().Str("URL", fmt.Sprintf("%s/app/%s/assessment/%s", config.UIHost, buildResponse.Application, buildResponse.Ref)).Msg("Assessment URL")

	if config.PollForMinutes <= 0 {
		log.Info().Msg("Succeeded")
		err = w.Write(buildResponse)
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.PollForMinutes)*time.Minute)
	defer cancel()
	taskResponse, err := pollForResults(ctx, client, config.Group, buildResponse.Package, buildResponse.Platform, buildResponse.Task)
	if err != nil {
		return err
	}

	if config.FindingsArtifactPath != "" {
		err := writeFindings(ctx, client, buildResponse.Task, config.FindingsArtifactPath)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("ArtifactPath", config.FindingsArtifactPath).Msg("Failed to write findings artifact")
		}
	}

	if !isAboveMinimum(taskResponse, config.MinimumScore) {
		log.Debug().Any("Task", taskResponse).Msg("Task")
		if err := w.Write(taskResponse.JSON2XX); err != nil {
			return err
		}
		return fmt.Errorf("the score %.2f is less than the required minimum %d", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
	}

	log.Info().Msg("Succeeded")

	return w.Write(taskResponse.JSON2XX)
}
