package main

import (
	"flag"
	"fmt"
	"github.com/kr/pretty"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	REDMINE_LOGO  = "http://www.redmine.org/attachments/3462/redmine_fluid_icon.png"
	YELLOW        = "#FFD442"
	GREEN         = "#048A25"
	BLUE          = "#496686"
	ISSUE_PATTERN = "(#|__url__\\/issues\\/)(\\d+)"
)

var (
	issuePattern *regexp.Regexp
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

func patternInit() {
	url := viper.GetString("redmine.url")

	// add slashes
	url = strings.Replace(url, "://", "\\://", 1)
	url = strings.Replace(url, "/", "\\/", -1)
	url = strings.Replace(url, ".", "\\.", -1)

	p := strings.Replace(ISSUE_PATTERN, "__url__", url, 1)
	issuePattern = regexp.MustCompile(p)
}

func processSlackEvent(ev *slack.MessageEvent, wg sync.WaitGroup) {
	defer wg.Done()

	matches := issuePattern.FindAllStringSubmatch(ev.Text, -1)
	// for _, v := range matches {
	// 	if issue, err := Client.GetIssue(strings.ToUpper(strings.TrimSpace(v[1]))); err == nil {
	// 		sendMessage(issue, channel)
	// 	}
	// }
}

func listenSlack() {
	slackApi := slack.New(viper.GetString("slack.token"))
	slackApi.SetDebug(true)

	rtm := slackApi.NewRTM()
	go rtm.ManageConnection()

	var wg sync.WaitGroup

SlackLoop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if ev.SubType != "bot_message" {
					wg.Add(1)
					go processSlackEvent(ev, wg)
				}
			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials for Slack!\n")
				break SlackLoop
			default:
				// Ignore other events..
			}
		}
	}

	wg.Wait()
}

func main() {
	configInit()
	patternInit()

	listenSlack()
}
