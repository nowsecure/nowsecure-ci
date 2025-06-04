package run

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunCommand(ctx context.Context, v *viper.Viper) *cobra.Command {
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
		NewRunFileCommand(v),
		NewRunIdCommand(v),
		NewRunPackageCommand(ctx, v),
	)

	return runCmd
}

func pollForResults(ctx context.Context, client *platformapi.ClientWithResponses, packageName, platform string, task float64, minutes int) (*platformapi.GetAppPlatformPackageAssessmentTaskResponse, error) {
	zerolog.Ctx(ctx).Debug().Int("minutes", minutes).Msg("Beginning polling")

	count := 0

	for count <= minutes {
		count++
		resp, err := client.GetAppPlatformPackageAssessmentTaskWithResponse(
			ctx,
			platformapi.GetAppPlatformPackageAssessmentTaskParamsPlatform(platform),
			packageName,
			int(task),
			nil)

		if err != nil {
			return nil, err
		}

		zerolog.Ctx(ctx).Debug().Int("status", resp.StatusCode()).Msg("Received http response")

		// If the assessment is not found, there was likely an uncaught error in submission. No need to continue polling
		if resp.HTTPResponse.StatusCode >= 400 && resp.HTTPResponse.StatusCode < 500 {
			return nil, resp.JSON4XX
		}

		// If there's an API error, it seems prudent to continue polling
		if resp.HTTPResponse.StatusCode >= 500 {
			zerolog.Ctx(ctx).Warn().
				Any("LabRouteError", resp.JSON5XX).
				Msg("Received 5XX error from Nowsecure API")
			continue
		}

		zerolog.Ctx(ctx).Debug().Any("task status", resp.JSON2XX.TaskStatus).Msg("Successfully polled for task status")

		var completed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "completed"
		var failed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "failed"
		if resp.StatusCode() == 200 {
			if *resp.JSON2XX.TaskStatus == completed || *resp.JSON2XX.TaskStatus == failed {
				zerolog.Ctx(ctx).Debug().Msg("Task has completed")
				return resp, nil
			}
		}

		zerolog.Ctx(ctx).Debug().Msg("Sleeping...")

		time.Sleep(1 * time.Minute)
	}

	return nil, errors.New("assessment not completed")
}

func isAboveMinimum(taskResponse *platformapi.GetAppPlatformPackageAssessmentTaskResponse, threshold int) bool {
	return *taskResponse.JSON2XX.AdjustedScore >= float32(threshold)
}
