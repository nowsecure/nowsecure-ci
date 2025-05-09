package run

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRunCommand(v *viper.Viper) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run an assessment for a given application",
		Long:  ``,
	}
	var (
		analysisType   string
		pollForMinutes int
		minimumScore   int
		group         string
	)

	runCmd.PersistentFlags().StringVar(&analysisType, "analysis-type", "full", "One of: full, static, sbom")
	runCmd.PersistentFlags().IntVar(&pollForMinutes, "poll-for-minutes", 60, "polling max duration")
	runCmd.PersistentFlags().IntVar(&minimumScore, "minimum-score", 0, "score threshold below which we exit code 1")

	runCmd.AddCommand(
		NewRunFileCommand(v),
		NewRunIdCommand(v),
		NewRunPackageCommand(v),
	)

	return runCmd
}
