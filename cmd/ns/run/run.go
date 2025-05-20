package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunCommand(v *viper.Viper) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run an assessment for a given application",
		Long:  ``,
	}

	runCmd.PersistentFlags().String("analysis-type", "full", "One of: full, static, sbom")
	runCmd.PersistentFlags().Int("poll-for-minutes", 60, "polling max duration")
	runCmd.PersistentFlags().Int("minimum-score", 0, "score threshold below which we exit code 1")

	err1 := v.BindPFlag("analysis_type", runCmd.PersistentFlags().Lookup("analysis-type"))
	err2 := v.BindPFlag("poll_for_minutes", runCmd.PersistentFlags().Lookup("poll-for-minutes"))
	err3 := v.BindPFlag("minimum_score", runCmd.PersistentFlags().Lookup("minimum-score"))

	if errs := errors.Join(err1, err2, err3); errs != nil {
		fmt.Println(errs)
		os.Exit(1)
	}

	runCmd.AddCommand(
		NewRunFileCommand(v),
		NewRunIdCommand(v),
		NewRunPackageCommand(v),
	)

	return runCmd
}

func pollForResults(ctx context.Context, client *platformapi.ClientWithResponses, packageName, platform string, task float64, minutes int) (*platformapi.GetAppPlatformPackageAssessmentTaskResponse, error) {
	fmt.Println("Polling for ", minutes)

	if minutes > 0 {
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

			fmt.Println(resp.StatusCode(), "-", *resp.JSON2XX.TaskStatus)

			var completed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "completed"
			var failed platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus = "failed"
			if resp.StatusCode() == 200 {
				if *resp.JSON2XX.TaskStatus == completed || *resp.JSON2XX.TaskStatus == failed {
					fmt.Println("Task has completed")
					return resp, nil
				}
			}

			fmt.Println("Sleeping")

			time.Sleep(1 * time.Minute)
		}
	}
	return nil, errors.New("assessment not completed")
}

func isAboveMinimum(taskResponse *platformapi.GetAppPlatformPackageAssessmentTaskResponse, threshold int) bool {
	return *taskResponse.JSON2XX.AdjustedScore >= float32(threshold)
}
