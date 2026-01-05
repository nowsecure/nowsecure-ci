package run

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func GetTestConfig(t *testing.T, doer *platformapi.TestRequestDoer) *internal.RunConfig {
	host := "https://localhost:8080"
	client, err := platformapi.ClientFromConfig(platformapi.Config{
		Host:      host,
		UserAgent: "test/1.0",
		Token:     "token",
	}, doer)
	require.NoError(t, err)

	config := &internal.RunConfig{
		BaseConfig: internal.BaseConfig{
			APIHost:        host,
			UIHost:         "https://localhost:8081",
			PlatformClient: client,
			Group:          types.UUID{},
			LogLevel:       zerolog.DebugLevel,
			Output:         "",
			OutputFormat:   output.JSON,
		},
		AnalysisType:         "full",
		PollForMinutes:       0,
		PollingInterval:      time.Second,
		MinimumScore:         0,
		Platform:             "android",
		FindingsArtifactPath: "",
		ArtifactsDir:         "",
	}

	return config
}

func useSuccessfulBuild(doer *platformapi.TestRequestDoer, appId uuid.UUID, packageName, platform string) {
	uploadResponse := &platformapi.PostBuild2XX1{
		Application: &appId,
		Package:     packageName,
		Platform:    platform,
		Task:        12345.6,
		Ref:         appId,
	}

	uploadBody, _ := json.Marshal(uploadResponse)

	doer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "POST" && req.URL.Path == "/build"
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(uploadBody)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil)
}

type GetAssessmentResponse struct {
	Account          *types.UUID                                                         `json:"account"`
	AdjustedIssues   *interface{}                                                        `json:"adjusted_issues"`
	AdjustedScore    *float32                                                            `json:"adjusted_score"`
	Application      *types.UUID                                                         `json:"application"`
	AppstoreDownload *platformapi.GetAppPlatformPackageAssessmentTask2XXAppstoreDownload `json:"appstore_download"`
	Binary           *string                                                             `json:"binary"`
	Cancelled        bool                                                                `json:"cancelled"`
	Config           struct {
		Dynamic interface{} `json:"dynamic"`
		Static  interface{} `json:"static"`
	} `json:"config"`
	Created *time.Time  `json:"created"`
	Creator *types.UUID `json:"creator"`
	Events  struct {
		Dynamic []interface{} `json:"dynamic"`
	} `json:"events"`
	Favorite          *bool        `json:"favorite"`
	Group             *types.UUID  `json:"group"`
	IdentifiedVulnMap *interface{} `json:"identified_vuln_map"`
	Package           string       `json:"package"`
	Platform          string       `json:"platform"`
	Ref               types.UUID   `json:"ref"`
	Status            struct {
		Dynamic interface{} `json:"dynamic"`

		Static interface{} `json:"static"`
	} `json:"status"`
	Task          float32                                                       `json:"task"`
	TaskErrorCode *string                                                       `json:"task_error_code"`
	TaskStatus    *platformapi.GetAppPlatformPackageAssessmentTask2XXTaskStatus `json:"task_status"`
	Updated       *time.Time                                                    `json:"updated"`
}

func UseSuccessfulPolling(doer *platformapi.TestRequestDoer, pollingResponse *GetAssessmentResponse) {
	assessmentBody, _ := json.Marshal(pollingResponse)
	bodyReader := bytes.NewReader(assessmentBody)
	doer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && strings.Contains(req.URL.Path, "assessment")
	})).Run(func(args mock.Arguments) {
		bodyReader.Seek(0, 0)
	}).Return(&http.Response{
		StatusCode:    200,
		Body:          io.NopCloser(bodyReader),
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		ContentLength: int64(len(assessmentBody)),
	}, nil)
}

func UseFlakyPolling(doer *platformapi.TestRequestDoer, pollingResponse *GetAssessmentResponse) {
	count := 0

	errorBody, _ := json.Marshal(&platformapi.LabRouteError{
		Description: platformapi.Ptr("Some error"),
		Message:     platformapi.Ptr("Some message"),
		Name:        platformapi.Ptr("Some name"),
		Status:      platformapi.Ptr("500"),
	})
	response := http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewReader(errorBody)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	assessmentBody, _ := json.Marshal(pollingResponse)
	bodyReader := bytes.NewReader(assessmentBody)
	doer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && strings.Contains(req.URL.Path, "assessment")
	})).Run(func(args mock.Arguments) {
		bodyReader.Seek(0, 0)
		count++
		if count > 1 {
			response = http.Response{
				StatusCode:    200,
				Body:          io.NopCloser(bodyReader),
				Header:        http.Header{"Content-Type": []string{"application/json"}},
				ContentLength: int64(len(assessmentBody)),
			}
		}
	}).Return(&response, nil)
}

func useSuccessfulAppList(doer *platformapi.TestRequestDoer, appList []platformapi.LabApp) {
	appListBody, _ := json.Marshal(appList)
	doer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "GET" && req.URL.Path == "/app"
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(appListBody)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil)
}

type TriggerAssessmentResponse struct {
	Application types.UUID
	Package     string
	Platform    string
	Task        float64
	Ref         types.UUID
}

func useSuccessfulTriggerAssessment(doer *platformapi.TestRequestDoer, mockResponse *TriggerAssessmentResponse) {
	triggerResponseBody, _ := json.Marshal(mockResponse)
	doer.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "POST" && strings.Contains(req.URL.Path, "assessment")
	})).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(triggerResponseBody)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil)
}
