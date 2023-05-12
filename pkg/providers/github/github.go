// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
	"github.com/open-component-model/git-controller/pkg/providers/gogit"
)

const (
	tokenKey      = "token"
	providerType  = "github"
	defaultDomain = github.DefaultDomain
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

// constructAuthenticationOption will take the object and construct an authentication option.
// For now, only token secret is supported, this will be extended in the future.
func (c *Client) constructAuthenticationOption(ctx context.Context, obj mpasv1alpha1.Repository) (gitprovider.ClientOption, error) {
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

	return gitprovider.WithOAuth2Token(string(token)), nil
}

func (c *Client) CreatePullRequest(ctx context.Context, branch string, sync deliveryv1alpha1.Sync, repository mpasv1alpha1.Repository) error {
	if repository.Spec.Provider != providerType {
		if c.next == nil {
			return fmt.Errorf("can't handle provider type '%s' and no next provider is configured", repository.Spec.Provider)
		}

		return c.next.CreatePullRequest(ctx, branch, sync, repository)
	}

	authenticationOption, err := c.constructAuthenticationOption(ctx, repository)
	if err != nil {
		return err
	}

	domain := defaultDomain
	if repository.Spec.Domain != "" {
		domain = repository.Spec.Domain
	}

	gc, err := github.NewClient(authenticationOption, gitprovider.WithDomain(domain))
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	if repository.Spec.IsOrganization {
		return gogit.CreateOrganizationPullRequest(ctx, gc, domain, branch, sync.Spec.PullRequestTemplate, repository)
	}

	return gogit.CreateUserPullRequest(ctx, gc, domain, branch, sync.Spec.PullRequestTemplate, repository)
}
