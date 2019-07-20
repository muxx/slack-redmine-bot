package main

import (
	"flag"
	"fmt"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"os"
	"slack-redmine-bot/redmine"
	"slack-redmine-bot/slack"
)

func configInit() {
	viper.SetEnvPrefix("srb")

	var configPath string

	flag.StringVar(&configPath, "config", "", "Path of configuration file without name (name must be config.yml)")
	flag.Parse()

	if len(configPath) > 0 {
		viper.AddConfigPath(configPath)
	}

	configPath = os.Getenv("SRB_CONFIG")
	if len(configPath) > 0 {
		viper.AddConfigPath(configPath)
	}

	viper.AddConfigPath("/etc/slack-redmine-bot/")
	viper.AddConfigPath(".")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error in config file. %s \n", err))
	}
}

func init() {
	configInit()
}

func main() {
	redmine := redmine.New(
		viper.GetString("redmine.url"),
		viper.GetString("redmine.api_key"),
		cast.ToIntSlice(viper.Get("redmine.closed_statuses")),
		cast.ToIntSlice(viper.Get("redmine.high_priorities")),
	)
	slack := slack.New(redmine, viper.GetString("slack.token"))

	slack.Listen()
}
