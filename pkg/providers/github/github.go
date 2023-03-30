package github

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/open-component-model/git-controller/pkg/providers"
)

type Client struct {
	// TODO: Figure out how to get this.
	BaseURL string
}

func NewClient() *Client {
	return &Client{}
}

var _ providers.Provider = &Client{}

func (c *Client) CreateRepository(ctx context.Context, owner, repo string) error {
	_, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	return nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error {
	return nil
}
