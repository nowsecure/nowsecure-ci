package util

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Config struct {
	Host      string
	Token     string
	Group     string
	UserAgent string
}

func NewConfig(v *viper.Viper) Config {
	host := v.GetString("host")
	token := v.GetString("token")

	if host == "" || token == "" {
		fmt.Println("Host and token must both be specified either in a config file, or through a flag")
		os.Exit(1)
	}

	return Config{
		Host:      host,
		Token:     token,
		Group:     v.GetString("group"),
		UserAgent: v.GetString("userAgent"),
	}
}

func ViperWithFile(configFile string) *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := homedir.Dir()
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
