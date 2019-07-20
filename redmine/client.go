package redmine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	IssueUrl     = "%s/issues/%d"
	IssueUrlJson = "%s/issues/%s.json"
)

type Client struct {
	Url            string
	apiKey         string
	httpClient     *http.Client
	closedStatuses []int
	highPriorities []int
}

// temporary structure for json decoding
type IssueWrapper struct {
	Issue *Issue `json:"issue"`
}

type Issue struct {
	Id         int         `json:"id"`
	Subject    string      `json:"subject"`
	Status     *Dictionary `json:"status"`
	Project    *Dictionary `json:"project"`
	Tracker    *Dictionary `json:"tracker"`
	Priority   *Dictionary `json:"priority"`
	Author     *Dictionary `json:"author"`
	AssignedTo *Dictionary `json:"assigned_to"`
	Category   *Dictionary `json:"category"`
	Version    *Dictionary `json:"fixed_version"`
}

type Dictionary struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func New(url string, apiKey string, statuses []int, priorities []int) *Client {
	return &Client{
		url,
		apiKey,
		&http.Client{Timeout: 10 * time.Second},
		statuses,
		priorities,
	}
}

func (r *Client) GetIssueUrl(issue *Issue) string {
	return fmt.Sprintf(IssueUrl, r.Url, issue.Id)
}

func (r *Client) IssueIsClosed(issue *Issue) bool {
	if issue.Status == nil {
		return false
	}

	for _, s := range r.closedStatuses {
		if s == issue.Status.Id {
			return true
		}
	}

	return false
}

func (r *Client) IssueInHighPriority(issue *Issue) bool {
	if issue.Priority == nil {
		return false
	}

	for _, p := range r.highPriorities {
		if p == issue.Priority.Id {
			return true
		}
	}

	return false
}

func (r *Client) GetIssue(number string) (*Issue, error) {
	var issue Issue

	req, err := http.NewRequest("GET", fmt.Sprintf(IssueUrlJson, r.Url, number), nil)
	if err != nil {
		return &issue, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Add("X-Redmine-API-Key", r.apiKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return &issue, err
	}

	if resp.StatusCode != http.StatusOK {
		return &issue, errors.New(fmt.Sprintf("Error getting issue. Status code: %d.\n", resp.StatusCode))
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &issue, err
	}

	var iw IssueWrapper
	if err := json.Unmarshal(data, &iw); err != nil {
		return &issue, err
	} else {
		issue = *iw.Issue
	}

	return &issue, nil
}
