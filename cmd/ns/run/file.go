package run

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal/util"
)

func NewRunFileCommand(v *viper.Viper) *cobra.Command {
	// fileCmd represents the file command
	var fileCmd = &cobra.Command{
		Use:       "file [./file-path]",
		Short:     "Upload and run an assessment for a specified binary file",
		Long:      ``,
		ValidArgs: []string{"file"},
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			file := args[0]
			fmt.Println(util.NewConfig(v))
			fmt.Println("file called with file ", file)
		},
	}

	return fileCmd
}
