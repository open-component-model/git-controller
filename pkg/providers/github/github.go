// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	ggithub "github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
	"github.com/open-component-model/git-controller/pkg/providers/gogit"
)

const (
	tokenKey        = "password"
	providerType    = "github"
	defaultDomain   = github.DefaultDomain
	statusCheckName = "mpas/validation-check"
)

// Client github.
type Client struct {
	client client.Client
	next   providers.Provider
}

// NewClient creates a new GitHub client.
func NewClient(client client.Client, next providers.Provider) *Client {
	return &Client{
		client: client,
		next:   next,
	}
}

var _ providers.Provider = &Client{}

func (c *Client) CreateRepository(ctx context.Context, obj mpasv1alpha1.Repository) error {
	if obj.Spec.Provider != providerType {
		if c.next == nil {
			return fmt.Errorf("can't handle provider type '%s' and no next provider is configured", obj.Spec.Provider)
		}

		return c.next.CreateRepository(ctx, obj)
	}

	authenticationOption, err := c.constructAuthenticationOption(ctx, obj)
	if err != nil {
		return err
	}

	domain := defaultDomain
	if obj.Spec.Domain != "" {
		domain = obj.Spec.Domain
	}

	gc, err := github.NewClient(authenticationOption, gitprovider.WithDomain(domain))
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	if obj.Spec.IsOrganization {
		return gogit.CreateOrganizationRepository(ctx, gc, domain, obj)
	}

	return gogit.CreateUserRepository(ctx, gc, domain, obj)
}

func (c *Client) CreateBranchProtection(ctx context.Context, obj mpasv1alpha1.Repository) error {
	token, err := c.retrieveAccessToken(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to retrieve token: %w", err)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
	tc := oauth2.NewClient(context.Background(), ts)

	g := ggithub.NewClient(tc)

	if _, _, err := g.Repositories.UpdateBranchProtection(ctx, obj.Spec.Owner, obj.Name, obj.Spec.DefaultBranch, &ggithub.ProtectionRequest{
		RequiredStatusChecks: &ggithub.RequiredStatusChecks{
			Strict: true,
			Checks: []*ggithub.RequiredStatusCheck{
				{
					Context: statusCheckName,
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to update branch protection rules: %w", err)
	}

	return nil
}

// constructAuthenticationOption will take the object and construct an authentication option.
// For now, only token secret is supported, this will be extended in the future.
func (c *Client) constructAuthenticationOption(ctx context.Context, obj mpasv1alpha1.Repository) (gitprovider.ClientOption, error) {
	token, err := c.retrieveAccessToken(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	return gitprovider.WithOAuth2Token(string(token)), nil
}

func (c *Client) retrieveAccessToken(ctx context.Context, obj mpasv1alpha1.Repository) ([]byte, error) {
	secret := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      obj.Spec.Credentials.SecretRef.Name,
		Namespace: obj.Namespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("token '%s' not found in secret", tokenKey)
	}

	return token, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, branch string, sync deliveryv1alpha1.Sync, repository mpasv1alpha1.Repository) (int, error) {
	if repository.Spec.Provider != providerType {
		if c.next == nil {
			return -1, fmt.Errorf("can't handle provider type '%s' and no next provider is configured", repository.Spec.Provider)
		}

		return c.next.CreatePullRequest(ctx, branch, sync, repository)
	}

	authenticationOption, err := c.constructAuthenticationOption(ctx, repository)
	if err != nil {
		return -1, err
	}

	domain := defaultDomain
	if repository.Spec.Domain != "" {
		domain = repository.Spec.Domain
	}

	gc, err := github.NewClient(authenticationOption, gitprovider.WithDomain(domain))
	if err != nil {
		return -1, fmt.Errorf("failed to create github client: %w", err)
	}

	var (
		id                    int
		createPullRequestFunc = gogit.CreateUserPullRequest
	)

	if repository.Spec.IsOrganization {
		createPullRequestFunc = gogit.CreateOrganizationPullRequest
	}

	if id, err = createPullRequestFunc(ctx, gc, domain, branch, sync.Spec.PullRequestTemplate, repository); err != nil {
		return 0, fmt.Errorf("failed to create pull request: %w", err)
	}

	if err := c.createCheckRun(ctx, repository, id); err != nil {
		return 0, fmt.Errorf("failed to create check run: %w", err)
	}

	return id, nil
}

func (c *Client) createCheckRun(ctx context.Context, repository mpasv1alpha1.Repository, prID int) error {
	token, err := c.retrieveAccessToken(ctx, repository)
	if err != nil {
		return fmt.Errorf("failed to retrieve token: %w", err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
	tc := oauth2.NewClient(context.Background(), ts)

	g := ggithub.NewClient(tc)
	//
	pr, _, err := g.PullRequests.Get(ctx, repository.Spec.Owner, repository.Name, prID)
	if err != nil {
		return fmt.Errorf("failed to find PR: %w", err)
	}

	_, _, err = g.Repositories.CreateStatus(ctx, repository.Spec.Owner, repository.Name, *pr.Head.SHA, &ggithub.RepoStatus{
		State:       ggithub.String("pending"),
		Description: ggithub.String("MPAS Validation Check"),
		Context:     ggithub.String(statusCheckName),
	})

	if err != nil {
		return fmt.Errorf("failed to create status for pr: %w", err)
	}

	return nil
}
