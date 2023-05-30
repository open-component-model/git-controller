// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gitea

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"code.gitea.io/sdk/gitea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

const (
	tokenKey      = "password"
	providerType  = "gitea"
	defaultDomain = "gitea.com"
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

	domain := defaultDomain
	if obj.Spec.Domain != "" {
		domain = obj.Spec.Domain
	}

	client, err := gitea.NewClient(domain, gitea.SetToken(string(token)))
	if err != nil {
		return fmt.Errorf("failed to create gitea client: %w", err)
	}

	private := true
	if obj.Spec.Visibility == "public" {
		private = false
	}

	if _, _, err := client.CreateRepo(gitea.CreateRepoOption{
		Name:          obj.GetName(),
		Description:   "Created by git-controller",
		Private:       private,
		AutoInit:      true,
		DefaultBranch: "main",
		TrustModel:    gitea.TrustModelDefault,
	}); err != nil {
		return fmt.Errorf("failed to create repositroy: %w", err)
	}

	f := &fileCommitter{}

	if len(obj.Spec.Maintainers) != 0 {
		var content []byte
		buffer := bytes.NewBuffer(content)

		for _, m := range obj.Spec.Maintainers {
			buffer.WriteString(m)
		}

		encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())

		f.commitFile(client, obj, "CODEOWNERS", encoded)
	}

	f.commitFile(client, obj, "generators/.keep", "")
	f.commitFile(client, obj, "products/.keep", "")
	f.commitFile(client, obj, "subscriptions/.keep", "")
	f.commitFile(client, obj, "targets/.keep", "")

	if f.err != nil {
		return fmt.Errorf("failed to set up project folder structure: %w", f.err)
	}

	return nil
}

type fileCommitter struct {
	err error
}

func (f *fileCommitter) commitFile(client *gitea.Client, obj mpasv1alpha1.Repository, path, content string) {
	if f.err != nil {
		return
	}

	_, _, err := client.CreateFile(obj.Spec.Owner, obj.GetName(), path, gitea.CreateFileOptions{
		FileOptions: gitea.FileOptions{
			Message:    fmt.Sprintf("Adding '%s' file.", path),
			BranchName: obj.Spec.DefaultBranch,
		},
		Content: content,
	})
	if err != nil {
		if _, derr := client.DeleteRepo(obj.Spec.Owner, obj.GetName()); derr != nil {
			err = errors.Join(err, derr)
		}

		f.err = fmt.Errorf("failed to add file '%s' file: %w", path, err)
	}
}

func (c *Client) CreatePullRequest(ctx context.Context, branch string, sync deliveryv1alpha1.Sync, repository mpasv1alpha1.Repository) error {
	if repository.Spec.Provider != providerType {
		if c.next == nil {
			return fmt.Errorf("can't handle provider type '%s' and no next provider is configured", repository.Spec.Provider)
		}

		return c.next.CreatePullRequest(ctx, branch, sync, repository)
	}

	secret := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      repository.Spec.Credentials.SecretRef.Name,
		Namespace: repository.Namespace,
	}, secret); err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	token, ok := secret.Data[tokenKey]
	if !ok {
		return fmt.Errorf("token '%s' not found in secret", tokenKey)
	}

	domain := defaultDomain
	if repository.Spec.Domain != "" {
		domain = repository.Spec.Domain
	}

	client, err := gitea.NewClient(domain, gitea.SetToken(string(token)))
	if err != nil {
		return fmt.Errorf("failed to create gitea client: %w", err)
	}

	var (
		title       = providers.DefaultTitle
		base        = providers.DefaultBaseBranch
		description = providers.DefaultDescription
	)

	if sync.Spec.PullRequestTemplate.Title != "" {
		title = sync.Spec.PullRequestTemplate.Title
	}

	if sync.Spec.PullRequestTemplate.Base != "" {
		base = sync.Spec.PullRequestTemplate.Base
	}

	if sync.Spec.PullRequestTemplate.Description != "" {
		description = sync.Spec.PullRequestTemplate.Description
	}

	if _, _, err := client.CreatePullRequest(repository.Spec.Owner, repository.GetName(), gitea.CreatePullRequestOption{
		Head:  branch,
		Base:  base,
		Title: title,
		Body:  description,
	}); err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	return nil
}

func (c *Client) CreateBranchProtection(ctx context.Context, obj mpasv1alpha1.Repository) error {
	return providers.NotSupportedError
}
