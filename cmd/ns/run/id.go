package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	appId string
)

func NewRunIdCommand() *cobra.Command {
	// idCmd represents the id command
	var idCmd = &cobra.Command{
		Use:       "id [app-id]",
		Short:     "Run an assessment for a pre-existing app by specifying app-id",
		Long:      ``,
		ValidArgs: []string{"appId"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("package called")
		},
	}
	return idCmd

}
