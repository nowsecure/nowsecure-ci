package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	android     bool
	ios         bool
	packageName string
)

func NewRunPackageCommand() *cobra.Command {
	// packageCmd represents the package command
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			packageName = args[0]
			fmt.Println("package called with package ", packageName)
		},
	}

	packageCmd.Flags().BoolVar(&ios, "iOS", false, "app is for iOS platform")
	packageCmd.Flags().BoolVar(&android, "android", false, "app is for android platform")

	packageCmd.MarkFlagsOneRequired("iOS", "android")
	packageCmd.MarkFlagsMutuallyExclusive("iOS", "android")

	return packageCmd
}
