package run

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func TestByFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test.apk")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	packageName := "com.example"
	appId := uuid.New()
	completedStatus := platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus("completed")

	t.Run("Successful assessment without polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		appId := uuid.New()

		useSuccessfulBuild(t, doer, appId, packageName, config.Platform)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err = ByFile(ctx, tmpFile.Name(), config)
		assert.NoError(t, err)
	})

	t.Run("Successful assessment with polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 1

		useSuccessfulBuild(t, doer, appId, packageName, config.Platform)

		assessmentResponse := &GetAssessmentResponse{
			Application:   &appId,
			Package:       packageName,
			Platform:      config.Platform,
			Task:          1234.50,
			Ref:           appId,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(85.5)),
		}

		UseSuccessfulPolling(t, doer, assessmentResponse)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err = ByFile(ctx, tmpFile.Name(), config)
		assert.NoError(t, err)
	})

	t.Run("Successful assessment with flaky API", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 2

		useSuccessfulBuild(t, doer, appId, packageName, config.Platform)

		assessmentResponse := &GetAssessmentResponse{
			Application:   &appId,
			Package:       packageName,
			Platform:      config.Platform,
			Task:          1234.50,
			Ref:           appId,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(85.5)),
		}

		UseFlakyPolling(t, doer, assessmentResponse)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err = ByFile(ctx, tmpFile.Name(), config)
		assert.NoError(t, err)
	})

	t.Run("Assessment below minimum score throws an error", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 2
		config.MinimumScore = 50

		useSuccessfulBuild(t, doer, appId, packageName, config.Platform)

		assessmentResponse := &GetAssessmentResponse{
			Application:   &appId,
			Package:       packageName,
			Platform:      config.Platform,
			Task:          1234.50,
			Ref:           appId,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(5.5)),
		}

		UseSuccessfulPolling(t, doer, assessmentResponse)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err = ByFile(ctx, tmpFile.Name(), config)
		require.ErrorContains(t, err, "less than the required minimum")
	})

	t.Run("Assessment against missing file throws an error", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err = ByFile(ctx, "/nonexistent/file.apk", config)
		require.Error(t, err)
	})
}
