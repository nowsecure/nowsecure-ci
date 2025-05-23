package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRunPackageCommand() *cobra.Command {
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"package-name"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			packageName := args[0]
			fmt.Println("package called with package ", packageName)
		},
	}
	var (
		android bool
		ios     bool
	)

	packageCmd.Flags().BoolVar(&ios, "ios", false, "app is for ios platform")
	packageCmd.Flags().BoolVar(&android, "android", false, "app is for android platform")

	packageCmd.MarkFlagsOneRequired("ios", "android")
	packageCmd.MarkFlagsMutuallyExclusive("ios", "android")

	return packageCmd
}
