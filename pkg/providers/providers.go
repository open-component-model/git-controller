package providers

import (
	"context"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
)

// Provider adds the ability to create repositories and pull requests.
type Provider interface {
	CreateRepository(ctx context.Context, spec mpasv1alpha1.Repository) error
	CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error
}
