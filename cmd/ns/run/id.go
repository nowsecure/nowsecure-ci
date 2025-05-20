package run

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
)

func NewRunIdCommand(v *viper.Viper) *cobra.Command {
	// idCmd represents the id command
	var idCmd = &cobra.Command{
		Use:       "id [app-id]",
		Short:     "Run an assessment for a pre-existing app by specifying app-id",
		Long:      ``,
		ValidArgs: []string{"appId"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appId := args[0]
			config := internal.NewConfig(v)
			fmt.Println(config)
			fmt.Println("package called with ", appId)
		},
	}
	return idCmd
}
