package utils

import (
	"context"
	"log"

	"github.com/google/go-github/github"
)

// Configuration contains some sensitive passwords and usernames
type Configuration struct {
	GithubPersonalAccessToken string `json:"GithubPersonalAccessToken"`
	GithubOwner               string `json:"GithubOwner"`
	GithubRepo                string `json:"GithubRepo"`
}

// CreateGithubIssue creates a new Github issue
func CreateGithubIssue(ctx context.Context, conf *Configuration, client *github.Client, labels *[]string, title, body string) {
	issueTitle := title
	issueBody := body
	_, _, err := client.Issues.Create(ctx, conf.GithubOwner, conf.GithubRepo, &github.IssueRequest{
		Title:    &issueTitle,
		Body:     &issueBody,
		Labels:   labels,
		Assignee: &conf.GithubOwner,
	})
	if err != nil {
		log.Fatal(err)
	}
}
