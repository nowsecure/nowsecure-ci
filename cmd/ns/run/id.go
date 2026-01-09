package run

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func IDCommand(v *viper.Viper, config *internal.BaseConfig) *cobra.Command {
	var idCmd = &cobra.Command{
		Use:       "id [app-id]",
		Short:     "Run an assessment for a pre-existing app by specifying app-id",
		Long:      ``,
		ValidArgs: []string{"appId"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appID, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}
			config, err := internal.NewRunConfig(v)
			if err != nil {
				return err
			}
			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())

			return ByID(ctx, appID, config)
		},
	}
	return idCmd
}

func ByID(ctx context.Context, appID uuid.UUID, config *internal.RunConfig) error {
	log := zerolog.Ctx(ctx)
	client := config.PlatformClient

	w, err := output.New(config.Output, config.OutputFormat)
	if err != nil {
		return err
	}
	defer w.Close()

	appList, err := platformapi.GetAppList(ctx, client, platformapi.GetAppParams{
		Platform: (*platformapi.GetAppParamsPlatform)(&config.Platform),
		Package:  nil,
		Group:    &config.Group,
		Ref:      &appID,
	})
	if err != nil {
		return err
	}
	if len(appList) != 1 {
		return fmt.Errorf("got %d elements but expected exactly one", len(appList))
	}

	app := appList[0]
	config.Platform = string(app.Platform)

	response, err := platformapi.TriggerAssessment(ctx, client, platformapi.TriggerAssessmentParams{
		PackageName:  app.Package,
		Group:        config.Group,
		AnalysisType: config.AnalysisType,
		Platform:     string(app.Platform),
	})
	if err != nil {
		return err
	}
	log.Info().Str("URL", fmt.Sprintf("%s/app/%s/assessment/%s", config.UIHost, response.JSON2XX.Application, response.JSON2XX.Ref)).Msg("Assessment URL")

	if config.PollForMinutes <= 0 {
		log.Info().Msg("Succeeded")
		return w.Write(response.JSON2XX)
	}

	ticker := time.NewTicker(1 * config.PollingInterval)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.PollForMinutes)*config.PollingInterval)
	defer cancel()
	taskResponse, err := pollForResults(ctx, client, ticker, config.Group, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task))
	if err != nil {
		return err
	}

	if config.FindingsArtifactPath != "" {
		err := writeFindings(ctx, client, float64(response.JSON2XX.Task), config.FindingsArtifactPath)
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
