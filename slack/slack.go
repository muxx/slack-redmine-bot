package slack

import (
	"errors"
	"fmt"
	"github.com/muxx/slack-redmine-bot/redmine"
	slackapi "github.com/nlopes/slack"
	"regexp"
	"sync"
)

const (
	botLogo      = "http://139810.selcdn.com/download/redmine_fluid_icon.png"
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
	api := slackapi.New(token)
	api.SetDebug(true)

	return &Client{api, r, issuePattern}
}

func (s *Client) sendMessage(issue *redmine.Issue, channel string) (err error) {
	params := slackapi.PostMessageParameters{}
	params.IconURL = botLogo
	params.Username = "Redmine Bot"

	fields := make([]slackapi.AttachmentField, 6)
	var idx int = 3

	fields[0] = slackapi.AttachmentField{
		Title: "Project",
		Value: issue.Project.Name,
		Short: true,
	}
	fields[1] = slackapi.AttachmentField{
		Title: "Status",
		Value: issue.Status.Name,
		Short: true,
	}
	fields[2] = slackapi.AttachmentField{
		Title: "Author",
		Value: issue.Author.Name,
		Short: true,
	}
	if issue.AssignedTo != nil {
		fields[idx] = slackapi.AttachmentField{
			Title: "Assigned To",
			Value: issue.AssignedTo.Name,
			Short: true,
		}
		idx += 1
	}
	if issue.Category != nil {
		fields[idx] = slackapi.AttachmentField{
			Title: "Category",
			Value: issue.Category.Name,
			Short: true,
		}
		idx += 1
	}
	if issue.Version != nil {
		fields[idx] = slackapi.AttachmentField{
			Title: "Version",
			Value: issue.Version.Name,
			Short: true,
		}
	}

	var title string
	if issue.Tracker != nil {
		title = fmt.Sprintf("%s #%d: %s", issue.Tracker.Name, issue.Id, issue.Subject)
	} else {
		title = fmt.Sprintf("#%d: %s", issue.Id, issue.Subject)
	}

	attachment := slackapi.Attachment{
		Title:     title,
		TitleLink: s.redmine.GetIssueUrl(issue),
		Fields:    fields,
	}

	if s.redmine.IssueIsClosed(issue) {
		attachment.Color = "good"
	} else if s.redmine.IssueInHighPriority(issue) {
		attachment.Color = "danger"
	}

	params.Attachments = []slackapi.Attachment{attachment}

	_, _, err = s.slack.PostMessage(channel, "", params)

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

		err = s.sendMessage(issue, ev.Channel)
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
				if ev.SubType != "bot_message" && ev.SubType != "message_deleted" {
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
