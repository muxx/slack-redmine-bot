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
	api := slackapi.New(token, slackapi.OptionDebug(true))

	return &Client{api, r, issuePattern}
}

func (s *Client) sendMessage(issue *redmine.Issue, channel string, threadTimestamp string) (err error) {
	var options []slackapi.MsgOption

	params := slackapi.PostMessageParameters{
		IconURL:  botLogo,
		Username: "Redmine Bot",
	}
	if threadTimestamp != "" {
		params.ThreadTimestamp = threadTimestamp
	}

	options = append(options, slackapi.MsgOptionPostMessageParameters(params))

	fields := make([]slackapi.AttachmentField, 6)
	var idx = 3

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

	options = append(options, slackapi.MsgOptionAttachments(attachment))

	_, _, err = s.slack.PostMessage(channel, options...)

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
