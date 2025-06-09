package platformapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"

	"github.com/nowsecure/nowsecure-ci/internal"
)

type LoggingDoer struct {
	client *http.Client
}

func (ld *LoggingDoer) Do(req *http.Request) (*http.Response, error) {
	zerolog.Ctx(req.Context()).Debug().Str("Path", req.URL.Path).Str("Query", req.URL.RawQuery).Msg("Platform request")
	resp, err := ld.client.Do(req)
	zerolog.Ctx(req.Context()).Debug().Str("Status", resp.Status).Msg("Platform response")
	return resp, err
}

func ClientFromConfig(config *internal.RunConfig, doer HttpRequestDoer) (*ClientWithResponses, error) {
	if doer == nil {
		doer = &LoggingDoer{&http.Client{}}
	}

	return NewClientWithResponses(config.Host,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", config.UserAgent)
			req.Header.Add("Authorization", "Bearer "+config.Token)
			return nil
		}), WithHTTPClient(doer))
}

type TriggerAssessmentParams struct {
	PackageName  string
	Group        types.UUID
	AnalysisType string
	Platform     string
}

func TriggerAssessment(ctx context.Context, client *ClientWithResponses, p TriggerAssessmentParams) (*PostAppPlatformPackageAssessmentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Debug().Str("package", p.PackageName).Str("platform", p.Platform).Msg("Triggering assessment")
	response, err := client.PostAppPlatformPackageAssessmentWithResponse(
		ctx,
		PostAppPlatformPackageAssessmentParamsPlatform(p.Platform),
		p.PackageName,
		&PostAppPlatformPackageAssessmentParams{
			Group:                   &p.Group,
			AppstoreDownload:        nil,
			Failfast:                Ptr(true),
			AnalysisType:            (*PostAppPlatformPackageAssessmentParamsAnalysisType)(&p.AnalysisType),
			HideSensitiveDataValues: Ptr(false),
		},
	)
	if err != nil {
		return nil, err
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, response.JSON4XX
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, response.JSON5XX
	}

	return response, nil
}

type UploadFileParams struct {
	AnalysisType string
	Group        types.UUID
	File         *os.File
}

func UploadFile(ctx context.Context, client *ClientWithResponses, p UploadFileParams) (*PostBuild2XX1, error) {
	response, err := client.PostBuildWithBodyWithResponse(ctx, &PostBuildParams{
		AnalysisType:            (*PostBuildParamsAnalysisType)(&p.AnalysisType),
		Group:                   &p.Group,
		Assessment:              Ptr(true),
		Version:                 nil,
		HideSensitiveDataValues: Ptr(false),
	}, "application/octet-stream", p.File)
	if err != nil {
		return nil, err
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, response.JSON4XX
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, response.JSON5XX
	}

	// NOTE: if the assessment build param gets changed to 'false' then handle PostBuild2XX0 as well
	buildResponse := PostBuild2XX1{}

	err = json.Unmarshal(response.Body, &buildResponse)

	return &buildResponse, err
}

func GetAppList(ctx context.Context, client *ClientWithResponses, p GetAppParams) ([]LabApp, error) {
	response, err := client.GetAppWithResponse(ctx, &p)
	if err != nil {
		return nil, err
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, response.JSON4XX
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, response.JSON5XX
	}

	zerolog.Ctx(ctx).Debug().Any("response", response.JSON2XX).Msg("Get app response")
	return *response.JSON2XX, nil
}

type GetAssessmentParams struct {
	Platform    string
	PackageName string
	TaskId      float64
	Group       types.UUID
}

// TODO impl these on clientWithResponses
func GetAssessment(ctx context.Context, client *ClientWithResponses, p GetAssessmentParams) (*GetAppPlatformPackageAssessmentTaskResponse, error) {
	resp, err := client.GetAppPlatformPackageAssessmentTaskWithResponse(
		ctx,
		GetAppPlatformPackageAssessmentTaskParamsPlatform(p.Platform),
		p.PackageName,
		//TODO: This cast should really not have to happen
		int(p.TaskId),
		&GetAppPlatformPackageAssessmentTaskParams{Group: Ptr(p.Group.String())},
	)
	if err != nil {
		return nil, err
	}

	if resp.HTTPResponse.StatusCode >= 400 && resp.HTTPResponse.StatusCode < 500 {
		return nil, resp.JSON4XX
	}

	if resp.HTTPResponse.StatusCode >= 500 {
		return nil, resp.JSON5XX
	}

	return resp, nil
}
