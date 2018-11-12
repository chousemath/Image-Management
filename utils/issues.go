package utils

import (
	"context"
	"log"

	"github.com/google/go-github/github"
)

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
