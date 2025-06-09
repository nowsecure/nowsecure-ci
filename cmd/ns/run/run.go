package run

import (
	"context"
	"errors"
	"time"

	types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

//revive:disable:exported
func RunCommand(ctx context.Context, v *viper.Viper) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run an assessment for a given application",
		Long:  ``,
	}

	runCmd.PersistentFlags().String("analysis-type", "full", "One of: full, static, sbom")
	runCmd.PersistentFlags().Int("poll-for-minutes", 60, "polling max duration")
	runCmd.PersistentFlags().Int("minimum-score", 0, "score threshold below which we exit code 1")
	bindingErrors := []error{
		v.BindPFlag("analysis_type", runCmd.PersistentFlags().Lookup("analysis-type")),
		v.BindPFlag("poll_for_minutes", runCmd.PersistentFlags().Lookup("poll-for-minutes")),
		v.BindPFlag("minimum_score", runCmd.PersistentFlags().Lookup("minimum-score")),
	}
	if errs := errors.Join(bindingErrors...); errs != nil {
		zerolog.Ctx(ctx).Panic().Err(errs).Msg("Failed binding run level flags")
	}

	runCmd.AddCommand(
		FileCommand(v),
		IDCommand(v),
		PackageCommand(ctx, v),
	)

	return runCmd
}

func pollForResults(ctx context.Context, client *platformapi.ClientWithResponses, group types.UUID, packageName, platform string, task float64) (*platformapi.GetAppPlatformPackageAssessmentTaskResponse, error) {
	zerolog.Ctx(ctx).Debug().Msg("Polling started")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := platformapi.GetAssessment(ctx, client, platformapi.GetAssessmentParams{
				Platform:    platform,
				PackageName: packageName,
				TaskId:      task,
				Group:       group,
			})
			if err != nil {
				if labErr, ok := err.(*platformapi.LabRouteError); ok {
					zerolog.Ctx(ctx).Error().Any("LabRouteError", labErr).Msg("API Error Response")
				}
			}

			zerolog.Ctx(ctx).Debug().Msg("Polling complete")

			var completed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "completed"
			var failed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "failed"
			if resp.StatusCode() == 200 {
				if *resp.JSON2XX.TaskStatus == completed || *resp.JSON2XX.TaskStatus == failed {
					return resp, nil
				}
			}
		}
	}
}

func isAboveMinimum(taskResponse *platformapi.GetAppPlatformPackageAssessmentTaskResponse, threshold int) bool {
	return *taskResponse.JSON2XX.AdjustedScore >= float32(threshold)
}
