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

func IDCommand(v *viper.Viper) *cobra.Command {
	var idCmd = &cobra.Command{
		Use:       "id [app-id]",
		Short:     "Run an assessment for a pre-existing app by specifying app-id",
		Long:      ``,
		ValidArgs: []string{"appId"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appId, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}

			config, err := internal.NewRunConfig(v)
			if err != nil {
				return err
			}

			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())
			log := zerolog.Ctx(ctx)
			log.Info().Any("AppId", appId).Msg("Package command called")

			client, err := platformapi.ClientFromConfig(config, nil)
			if err != nil {
				return err
			}

			w, err := output.New(config.Output, config.OutputFormat)
			if err != nil {
				return err
			}
			defer w.Close()

			appList, err := platformapi.GetAppList(ctx, client, platformapi.GetAppParams{
				Platform: (*platformapi.GetAppParamsPlatform)(&config.Platform),
				Package:  nil,
				Group:    &config.Group,
				Ref:      &appId,
			})
			if err != nil {
				return err
			}
			if len(*appList) != 1 {
				return fmt.Errorf("got %d elements but expected exactly one", len(*appList))
			}

			app := (*appList)[0]
			config.Platform = string(app.Platform)

			log.Info().Msg(config.Platform)

			response, err := platformapi.TriggerAssessment(ctx, client, platformapi.TriggerAssessmentParams{
				PackageName:  app.Package,
				Group:        config.Group,
				AnalysisType: config.AnalysisType,
				Platform:     string(app.Platform),
			})
			if err != nil {
				return err
			}

			if config.PollForMinutes <= 0 {
				log.Info().Msg("Succeeded")
				return w.Write(response.JSON2XX)
			}

			ctx, cancel := context.WithTimeout(ctx, time.Duration(config.PollForMinutes)*time.Minute)
			defer cancel()
			taskResponse, err := pollForResults(ctx, client, config.Group, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task), config.PollForMinutes)
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
	return idCmd
}
