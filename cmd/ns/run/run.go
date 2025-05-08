package run

import (
	"github.com/spf13/cobra"
)

var (
	analysisType   string
	pollForMinutes int
	minimumScore   int
	group          string
)

func NewRunCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run an assessment for a given application",
		Long:  ``,
	}

	runCmd.PersistentFlags().StringVar(&analysisType, "analysis-type", "full", "One of: full, static, sbom")
	runCmd.PersistentFlags().IntVar(&pollForMinutes, "poll-for-minutes", 60, "polling max duration")
	runCmd.PersistentFlags().IntVar(&minimumScore, "minimum-score", 0, "score threshold below which we exit code 1")
	runCmd.PersistentFlags().StringVarP(&group, "group", "g", "", "group with which to run assessment")

	runCmd.AddCommand(
		NewRunFileCommand(),
		NewRunIdCommand(),
		NewRunPackageCommand(),
	)

	return runCmd
}
