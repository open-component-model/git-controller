// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gitea

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

const (
	tokenKey     = "token"
	providerType = "gitea"
)

// Client gitea.
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

	secret := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      obj.Spec.Credentials.SecretRef.Name,
		Namespace: obj.Namespace,
	}, secret); err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	token, ok := secret.Data[tokenKey]
	if !ok {
		return fmt.Errorf("token '%s' not found in secret", tokenKey)
	}

	client, err := gitea.NewClient(obj.Spec.Domain, gitea.SetToken(string(token)))
	if err != nil {
		return fmt.Errorf("failed to create gitea client: %w", err)
	}

	private := true
	if obj.Spec.Visibility == "public" {
		private = false
	}

	if _, _, err := client.CreateRepo(gitea.CreateRepoOption{
		Name:          obj.Spec.RepositoryName,
		Description:   "Created by git-controller",
		Private:       private,
		AutoInit:      true,
		DefaultBranch: "main",
		TrustModel:    gitea.TrustModelDefault,
	}); err != nil {
		return fmt.Errorf("failed to create repositroy: %w", err)
	}

	return nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error {
	return nil
}
