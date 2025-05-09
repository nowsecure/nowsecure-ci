package run

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/util"
)

func NewRunPackageCommand(v *viper.Viper) *cobra.Command {
	// packageCmd represents the package command
	var packageCmd = &cobra.Command{
		Use:       "package [package-name]",
		Short:     "Run an assessment for a pre-existing app by specifying package and platform",
		Long:      ``,
		ValidArgs: []string{"packageName"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			packageName := args[0]
			config := util.NewConfig(v)
			fmt.Println(config)
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
