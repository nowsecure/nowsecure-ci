package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRunIdCommand() *cobra.Command {
	var idCmd = &cobra.Command{
		Use:       "id [app-id]",
		Short:     "Run an assessment for a pre-existing app by specifying app-id",
		Long:      ``,
		ValidArgs: []string{"appId"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appId := args[0]
			fmt.Println("package called with ", appId)
		},
	}
	return idCmd
}
