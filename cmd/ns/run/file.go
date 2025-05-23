package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRunFileCommand() *cobra.Command {
	var fileCmd = &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]
			fmt.Println("file called with file ", file)
		},
	}

	return fileCmd
}
