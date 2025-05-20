package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func NewRunFileCommand(v *viper.Viper) *cobra.Command {
	// fileCmd represents the file command
	var fileCmd = &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			fileName := args[0]
			file, err := os.Open(fileName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			config, configErr := internal.NewRunConfig(v)

			if configErr != nil {
				fmt.Println(configErr)
				os.Exit(1)
			}

			client, clientErr := internal.ClientFromConfig(config, nil)

			if clientErr != nil {
				fmt.Println(clientErr)
				os.Exit(1)
			}

			buildResponse, buildErr := submitFile(ctx, file, config, client)

			if buildErr != nil {
				fmt.Println(buildErr)
				os.Exit(1)
			}

			if config.PollForMinutes <= 0 {
				fmt.Println(buildResponse)
				return
			}

			taskResponse, taskErr := pollForResults(ctx, client, buildResponse.Package, buildResponse.Platform, buildResponse.Task, config.PollForMinutes)

			if taskErr != nil {
				fmt.Println(taskErr)
				os.Exit(1)
			}

			if config.MinimumScore <= 0 {
				fmt.Println(buildResponse)
				return
			}

			if !isAboveMinimum(taskResponse, config.MinimumScore) {
				fmt.Printf("The score %.2f is less than the required minimum %d\n", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
				os.Exit(1)
			}

			fmt.Println(buildResponse)
		},
	}

	return fileCmd
}

func submitFile(ctx context.Context, file *os.File, config internal.RunConfig, client *platformapi.ClientWithResponses) (*platformapi.PostBuild2XX1, error) {
	response, responseError := client.PostBuildWithBodyWithResponse(ctx, &platformapi.PostBuildParams{
		AnalysisType: (*platformapi.PostBuildParamsAnalysisType)(&config.AnalysisType),
		Group:        &config.Group,
	}, "application/octet-stream", file)

	if responseError != nil {
		fmt.Println(responseError)
	}

	if response.HTTPResponse.StatusCode >= 400 && response.HTTPResponse.StatusCode < 500 {
		return nil, errors.New(*response.JSON4XX.Description)
	}

	if response.HTTPResponse.StatusCode >= 500 {
		return nil, errors.New(*response.JSON5XX.Description)
	}

	buildResponse := platformapi.PostBuild2XX1{}

	err := json.Unmarshal(response.Body, &buildResponse)

	return &buildResponse, err
}
