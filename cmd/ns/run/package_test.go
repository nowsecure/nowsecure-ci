package run

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func TestByPackage(t *testing.T) {
	appID := uuid.New()
	packageName := "com.example"
	completedStatus := platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus("completed")

	t.Run("Successful assessment without polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    "ios",
			Task:        12345,
			Ref:         appID,
		})

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByPackage(ctx, packageName, config)
		assert.NoError(t, err)
	})

	t.Run("Successful assessment with polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 1
		config.MinimumScore = 85

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    "ios",
			Task:        12345,
			Ref:         appID,
		})

		UseSuccessfulPolling(t, doer, &GetAssessmentResponse{
			Application:   &appID,
			Package:       "com.example.iosapp",
			Platform:      "ios",
			Task:          12345,
			Ref:           appID,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(85.5)),
		})

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByPackage(ctx, packageName, config)
		assert.NoError(t, err)
	})

	t.Run("Successful assessment with flaky polling", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.PollForMinutes = 1
		config.MinimumScore = 85

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    "ios",
			Task:        12345,
			Ref:         appID,
		})

		pollingResponse := &GetAssessmentResponse{
			Application:   &appID,
			Package:       "com.example.iosapp",
			Platform:      "ios",
			Task:          12345,
			Ref:           appID,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(85.5)),
		}

		UseFlakyPolling(t, doer, pollingResponse)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByPackage(ctx, packageName, config)
		assert.NoError(t, err)
	})

	t.Run("Assessment below minimum score", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		config.AnalysisType = "static"
		config.PollForMinutes = 1
		config.MinimumScore = 75
		config.Platform = "ios"

		useSuccessfulTriggerAssessment(t, doer, &TriggerAssessmentResponse{
			Application: appID,
			Package:     packageName,
			Platform:    "ios",
			Task:        12345,
			Ref:         appID,
		})

		pollingResponse := &GetAssessmentResponse{
			Application:   &appID,
			Package:       "com.example.iosapp",
			Platform:      "ios",
			Task:          12345,
			Ref:           appID,
			TaskStatus:    &completedStatus,
			AdjustedScore: platformapi.Ptr(float32(25.5)),
		}

		UseSuccessfulPolling(t, doer, pollingResponse)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByPackage(ctx, "com.example.iosapp", config)
		require.ErrorContains(t, err, "less than the required minimum")
	})

	t.Run("Trigger assessment error", func(t *testing.T) {
		doer := &platformapi.TestRequestDoer{}
		config := GetTestConfig(t, doer)
		errorResponse := &platformapi.LabRouteError{
			Status:  platformapi.Ptr("404"),
			Name:    platformapi.Ptr("NotFound"),
			Message: platformapi.Ptr("Package not found"),
		}

		errorBody, _ := json.Marshal(errorResponse)
		doer.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewReader(errorBody)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil)

		ctx := zerolog.New(os.Stdout).WithContext(context.Background())
		err := ByPackage(ctx, "com.nonexistent.app", config)
		require.Error(t, err)
	})
}
