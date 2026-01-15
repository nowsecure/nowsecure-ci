package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
	"github.com/nowsecure/nowsecure-ci/internal/output"
	"github.com/nowsecure/nowsecure-ci/internal/platformapi"
)

func PackageCommand(c context.Context, v *viper.Viper, config *internal.BaseConfig) *cobra.Command {
	var packageCmd = &cobra.Command{
		Use:   "package [package-name]",
		Short: "Run an assessment for a pre-existing app by specifying package and platform",
		Example: `# Common flags
ns run package [package-name] \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --analysis-type static \
  --poll-for-minutes 30

# Run an assessment without waiting for results
ns run package [package-name] \
  --android \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 0

# Run a full (dynamic and static) assessment
ns run package [package-name] \
  --android \
  --analysis-type full \
  --group-ref YOUR_GROUP_UUID \
  --poll-for-minutes 60

# Run an assessment with a score threshold
ns run package [package-name] \
  --android \
  --analysis-type static \
  --minimum-score 70 \
  --poll-for-minutes 60 \
  --group-ref YOUR_GROUP_UUID
`,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, _ := internal.NewRunConfig(v)
			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())
			packageName := args[0]
			return ByPackage(ctx, packageName, config)
		},
	}

	packageCmd.Flags().Bool("ios", false, "app is for ios platform")
	packageCmd.Flags().Bool("android", false, "app is for android platform")

	packageCmd.MarkFlagsOneRequired("ios", "android")
	packageCmd.MarkFlagsMutuallyExclusive("ios", "android")

	bindingErrors := []error{
		v.BindPFlag("platform_android", packageCmd.Flags().Lookup("android")),
		v.BindPFlag("platform_ios", packageCmd.Flags().Lookup("ios")),
	}

	if errs := errors.Join(bindingErrors...); errs != nil {
		zerolog.Ctx(c).Panic().Err(errs).Msg("Failed binding run level flags")
	}

	return packageCmd
}

func ByPackage(ctx context.Context, packageName string, config *internal.RunConfig) error {
	log := zerolog.Ctx(ctx)
	w, err := output.New(config.Output, config.OutputFormat)
	if err != nil {
		return err
	}
	defer w.Close()

	client := config.PlatformClient

	response, err := platformapi.TriggerAssessment(ctx, client, platformapi.TriggerAssessmentParams{
		PackageName:  packageName,
		Group:        config.Group,
		AnalysisType: config.AnalysisType,
		Platform:     config.Platform,
	})
	if err != nil {
		return err
	}
	log.Info().Str("URL", fmt.Sprintf("%s/app/%s/assessment/%s", config.UIHost, response.JSON2XX.Application, response.JSON2XX.Ref)).Msg("Assessment URL")

	if config.PollForMinutes <= 0 {
		log.Info().Msg("Succeeded")
		return w.Write(response.JSON2XX)
	}

	ticker := time.NewTicker(1 * config.PollingInterval)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.PollForMinutes)*config.PollingInterval)
	defer cancel()
	taskResponse, err := pollForResults(ctx, client, ticker, config.Group, response.JSON2XX.Package, response.JSON2XX.Platform, float64(response.JSON2XX.Task))
	if err != nil {
		return err
	}

	if config.FindingsArtifactPath != "" {
		err := writeFindings(ctx, client, float64(response.JSON2XX.Task), config.FindingsArtifactPath)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("ArtifactPath", config.FindingsArtifactPath).Msg("Failed to write findings artifact")
			return err
		}
	}

	if !isAboveMinimum(taskResponse, config.MinimumScore) {
		log.Debug().Any("Task", taskResponse).Msg("Task")
		if err := w.Write(taskResponse.JSON2XX); err != nil {
			return err
		}
		return fmt.Errorf("the score %.2f is less than the required minimum %d", *taskResponse.JSON2XX.AdjustedScore, config.MinimumScore)
	}

	log.Info().Msg("Succeeded")
	return w.Write(taskResponse.JSON2XX)
}
