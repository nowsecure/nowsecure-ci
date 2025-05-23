package run

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/internal"
)

var (
	appId string
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
			config, _ := internal.NewRunConfig(v)
			ctx := internal.LoggerWithLevel(config.LogLevel).
				WithContext(cmd.Context())
			appId = args[0]
			zerolog.Ctx(ctx).Info().Str("AppId", appId).Msg("Package command called")
		},
	}
	return idCmd
}
