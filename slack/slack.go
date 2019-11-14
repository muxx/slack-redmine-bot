package slack

import (
	"errors"
	"fmt"
	slackapi "github.com/nlopes/slack"
	"regexp"
	"slack-redmine-bot/redmine"
	"sync"
)

const (
	botLogo      = "http://download.retailcrm.pro/redmine_fluid_icon.png"
	issuePattern = "(#|%s\\/issues\\/)(\\d+)"
)

type Client struct {
	slack   *slackapi.Client
	redmine *redmine.Client
	pattern *regexp.Regexp
}

func New(r *redmine.Client, token string) *Client {
	// init regexp
	p := fmt.Sprintf(issuePattern, regexp.QuoteMeta(r.Url))
	issuePattern := regexp.MustCompile(p)

	// Slack API client
	api := slackapi.New(token, slackapi.OptionDebug(true))

	return &Client{api, r, issuePattern}
}

func (s *Client) sendMessage(issue *redmine.Issue, channel string, threadTimestamp string) (err error) {
	// params
	params := slackapi.PostMessageParameters{
		IconURL:  botLogo,
		Username: "Redmine Bot",
	}
	if threadTimestamp != "" {
		params.ThreadTimestamp = threadTimestamp
	}

	// title
	var title string
	if issue.Tracker != nil {
		title = fmt.Sprintf("%s #%d: %s", issue.Tracker.Name, issue.Id, issue.Subject)
	} else {
		title = fmt.Sprintf("#%d: %s", issue.Id, issue.Subject)
	}
	titleText := slackapi.NewTextBlockObject(
		"mrkdwn",
		"*<"+s.redmine.GetIssueUrl(issue)+"|"+title+">*",
		false,
		false,
	)
	titleSection := slackapi.NewSectionBlock(titleText, nil, nil)

	// context
	elements := []slackapi.MixedElement{
		slackapi.NewTextBlockObject("mrkdwn", "*Project:* "+issue.Project.Name, false, false),
		slackapi.NewTextBlockObject("mrkdwn", "*Status:* "+issue.Status.Name, false, false),
	}

	if issue.Category != nil {
		elements = append(elements, slackapi.NewTextBlockObject("mrkdwn", "*Category:* "+issue.Category.Name, false, false))
	}
	if issue.Version != nil {
		elements = append(elements, slackapi.NewTextBlockObject("mrkdwn", "*Version:* "+issue.Version.Name, false, false))
	}
	if issue.AssignedTo != nil {
		elements = append(elements, slackapi.NewTextBlockObject("mrkdwn", "*Assigned To:* "+issue.AssignedTo.Name, false, false))
	}
	if s.redmine.IssueInHighPriority(issue) {
		elements = append(elements, slackapi.NewTextBlockObject("mrkdwn", "*Priority:* "+issue.Priority.Name, false, false))
	}

	contextSection := slackapi.NewContextBlock("", elements...)

	// send
	_, _, err = s.slack.PostMessage(
		channel,
		slackapi.MsgOptionPostMessageParameters(params),
		slackapi.MsgOptionBlocks(titleSection, contextSection),
	)

	return err
}

func (s *Client) processEvent(ev *slackapi.MessageEvent, wg sync.WaitGroup) {
	defer wg.Done()

	matches := s.pattern.FindAllStringSubmatch(ev.Text, -1)
	for _, v := range matches {
		issue, err := s.redmine.GetIssue(v[2])

		if err != nil {
			fmt.Printf("Error of the issue fetching. %s\n", err)
			continue
		}

		err = s.sendMessage(issue, ev.Channel, ev.ThreadTimestamp)

		if err != nil {
			fmt.Printf("Error of message sending. %s\n", err)
		}
	}
}

func (s *Client) Listen() {
	rtm := s.slack.NewRTM()
	go rtm.ManageConnection()

	var wg sync.WaitGroup

SlackLoop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slackapi.MessageEvent:
				if ev.SubType == "" || ev.SubType == "file_comment" || ev.SubType == "file_mention" {
					wg.Add(1)
					go s.processEvent(ev, wg)
				}
			case *slackapi.InvalidAuthEvent:
				fmt.Println(errors.New("Invalid credentials for Slack!\n"))
				break SlackLoop
			default:
				// Ignore other events..
			}
		}
	}

	wg.Wait()
}
