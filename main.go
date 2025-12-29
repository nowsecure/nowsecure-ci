package main

import (
	cmd "github.com/nowsecure/nowsecure-ci/cmd/ns"

	"context"
	"github.com/nowsecure/nowsecure-ci/internal"
	nserrors "github.com/nowsecure/nowsecure-ci/internal/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
)

func main() {
	ctx := zerolog.New(internal.ConsoleLevelWriter{}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.WarnLevel).
		WithContext(context.Background())

	v := viper.New()
	config := internal.BaseConfig{}

	root := cmd.RootCommand(ctx, v, &config)

	if err := root.ExecuteContext(ctx); err != nil {
		if reqErr, ok := err.(nserrors.CIError); ok {
			zerolog.Ctx(ctx).Error().Any("LabRouteError", reqErr).Msg("API Error Response")
			os.Exit(reqErr.ExitCode())
		}
		zerolog.Ctx(ctx).Fatal().Msg(err.Error())
	}
}
