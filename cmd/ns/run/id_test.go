package run

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func TestByID(t *testing.T) {
	appID := uuid.New()
	packageName := "com.example.app"
	completedStatus := platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus("completed")

	t.Run("Successful assessment without polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)

		useSuccessfulAppList(t, doer, []platformapi.LabApp{
			{
				Package:  packageName,
				Platform: "android",
			},
		})

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    config.Platform,
			Task:        12345,
			Ref:         appID,
		})

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByID(ctx, appID, config)
		require.NoError(t, err)
	})

	t.Run("Successful assessment with polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.AnalysisType = "sbom"
		config.PollForMinutes = 1
		config.MinimumScore = 90

		useSuccessfulAppList(t, doer, []platformapi.LabApp{
			{
				Package:  packageName,
				Platform: "android",
			},
		})

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    config.Platform,
			Task:        12345,
			Ref:         appID,
		})

		UseSuccessfulPolling(t, doer, &GetAssessmentResponse{
			Application:   &appID,
			Package:       packageName,
			Platform:      config.Platform,
			Task:          12345.50,
			Ref:           appID,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(92.5)),
		})

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByID(ctx, appID, config)
		require.NoError(t, err)
	})

	t.Run("App not found", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		appList := []platformapi.LabApp{}
		useSuccessfulAppList(t, doer, appList)
		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByID(ctx, appID, config)
		require.ErrorContains(t, err, "got 0 elements but expected exactly one")
	})

	t.Run("Multiple apps returned", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		appList := []platformapi.LabApp{
			{Package: "com.example.app1", Platform: "android"},
			{Package: "com.example.app2", Platform: "ios"},
		}

		useSuccessfulAppList(t, doer, appList)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByID(ctx, appID, config)
		require.ErrorContains(t, err, "got 2 elements but expected exactly one")
	})

	t.Run("Assessment below minimum score", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 1
		config.MinimumScore = 70

		appList := []platformapi.LabApp{
			{Package: "com.example.app", Platform: "android"},
		}
		useSuccessfulAppList(t, doer, appList)

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    config.Platform,
			Task:        12345,
			Ref:         appID,
		})

		UseSuccessfulPolling(t, doer, &GetAssessmentResponse{
			Application:   &appID,
			Package:       packageName,
			Platform:      config.Platform,
			Task:          12345,
			Ref:           appID,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(5.5)),
		})

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByID(ctx, appID, config)
		require.ErrorContains(t, err, "less than the required minimum")
	})
}
