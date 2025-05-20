package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nowsecure/nowsecure-ci/cmd/ns/run"
)

var rootCmd = &cobra.Command{
	Use:   "ns",
	Short: "NowSecure command line tool to interact with NowSecure Platform",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	configFile := ""
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.ns-ci)")

	v := viperWithFile(configFile)

	rootCmd.PersistentFlags().String("host", "https://lab-api.nowsecure.com", "REST API base url")
	rootCmd.PersistentFlags().String("token", "", "auth token for REST API")
	rootCmd.PersistentFlags().StringP("group", "g", "", "group with which to run assessments")

	v.SetDefault("userAgent", "nowsecure-ci")
	err1 := v.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	err2 := v.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	err3 := v.BindPFlag("group", rootCmd.PersistentFlags().Lookup("group"))

	if errs := errors.Join(err1, err2, err3); errs != nil {
		fmt.Println(errs)
		os.Exit(1)
	}

	rootCmd.AddCommand(run.NewRunCommand(v))
}

func viperWithFile(configFile string) *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defaultName := ".ns-ci"

		if _, err := os.Stat(filepath.Join(home, defaultName)); err != nil {
			return v
		}

		v.AddConfigPath(home)
		v.SetConfigName(defaultName)
		v.SetConfigType("yaml")
	}

	fmt.Println(v.ConfigFileUsed())

	err := v.ReadInConfig()
	if err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}

	return v
}
