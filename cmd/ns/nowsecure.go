package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/run"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ns",
	Short: "NowSecure command line tool to interact with NowSecure Platform",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	rootCmd.AddCommand(run.NewRunCommand())
}
