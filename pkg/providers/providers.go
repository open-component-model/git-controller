package providers

import "context"

type Provider interface {
	CreateRepository(ctx context.Context, owner, repo string) error
	CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error
}
