// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gogit

import (
	"context"
	"fmt"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CreateOrganizationRepository creates a repository for an authenticated organization.
func CreateOrganizationRepository(ctx context.Context, gc gitprovider.Client, domain string, spec mpasv1alpha1.RepositorySpec) error {
	logger := log.FromContext(ctx)

	visibility := gitprovider.RepositoryVisibility(spec.Visibility)

	if err := gitprovider.ValidateRepositoryVisibility(visibility); err != nil {
		return fmt.Errorf("failed to validate visibility: %w", err)
	}

	ref := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: spec.Owner,
		},
		RepositoryName: spec.RepositoryName,
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("main"),
		Visibility:    &visibility,
	}

	createOpts, err := gitprovider.MakeRepositoryCreateOptions(&gitprovider.RepositoryCreateOptions{AutoInit: gitprovider.BoolVar(true)})
	if err != nil {
		return fmt.Errorf("failed to create _create_ options for repository: %w", err)
	}

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		repo, err := gc.OrgRepositories().Create(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		if err := createCodeownersFile(ctx, repo, spec.Maintainers); err != nil {
			return fmt.Errorf("failed to add CODEOWNERS file: %w", err)
		}

		logger.Info("successfully created organization repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		repo, created, err := gc.OrgRepositories().Reconcile(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
			if err := createCodeownersFile(ctx, repo, spec.Maintainers); err != nil {
				return fmt.Errorf("failed to add CODEOWNERS file: %w", err)
			}

			logger.Info("successfully created organization repository", "domain", domain, "repository", spec.RepositoryName)
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", spec.ExistingRepositoryPolicy)
	}

	return nil
}

// CreateUserRepository creates a repository for an authenticated user.
func CreateUserRepository(ctx context.Context, gc gitprovider.Client, domain string, spec mpasv1alpha1.RepositorySpec) error {
	logger := log.FromContext(ctx)

	visibility := gitprovider.RepositoryVisibility(spec.Visibility)

	if err := gitprovider.ValidateRepositoryVisibility(visibility); err != nil {
		return fmt.Errorf("failed to validate visibility: %w", err)
	}

	ref := gitprovider.UserRepositoryRef{
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: spec.Owner,
		},
		RepositoryName: spec.RepositoryName,
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("main"),
		Visibility:    &visibility,
	}

	createOpts, err := gitprovider.MakeRepositoryCreateOptions(&gitprovider.RepositoryCreateOptions{AutoInit: gitprovider.BoolVar(true)})
	if err != nil {
		return fmt.Errorf("failed to create _create_ options for repository: %w", err)
	}

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		repo, err := gc.UserRepositories().Create(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		if err := createCodeownersFile(ctx, repo, spec.Maintainers); err != nil {
			return fmt.Errorf("failed to add CODEOWNERS file: %w", err)
		}

		logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		repo, created, err := gc.UserRepositories().Reconcile(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
			if err := createCodeownersFile(ctx, repo, spec.Maintainers); err != nil {
				return fmt.Errorf("failed to add CODEOWNERS file: %w", err)
			}

			logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", spec.ExistingRepositoryPolicy)
	}

	return nil
}

// CreateOrganizationPullRequest creates a pull-request for an organization owned repository.
func CreateOrganizationPullRequest(ctx context.Context, gc gitprovider.Client, domain, branch string, spec deliveryv1alpha1.PullRequestTemplate, repository mpasv1alpha1.RepositorySpec) error {
	// find the repository
	repo, err := gc.OrgRepositories().Get(ctx, gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: repository.Owner,
		},
		RepositoryName: repository.RepositoryName,
	})
	if err != nil {
		return fmt.Errorf("failed to find organization repository: %w", err)
	}

	var (
		title       = providers.DefaultTitle
		base        = providers.DefaultBaseBranch
		description = providers.DefaultDescription
	)

	if spec.Title != "" {
		title = spec.Title
	}

	if spec.Base != "" {
		base = spec.Base
	}

	if spec.Description != "" {
		description = spec.Description
	}

	pr, err := repo.PullRequests().Create(ctx, title, branch, base, description)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("created pull request for organization repository", "organization", repository.Owner, "pull-request", pr.Get().Number)

	return nil
}

// CreateUserPullRequest creates a pull-request for a user owned repository.
func CreateUserPullRequest(ctx context.Context, gc gitprovider.Client, domain, branch string, spec deliveryv1alpha1.PullRequestTemplate, repository mpasv1alpha1.RepositorySpec) error {
	// find the repository
	repo, err := gc.UserRepositories().Get(ctx, gitprovider.UserRepositoryRef{
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: repository.Owner,
		},
		RepositoryName: repository.RepositoryName,
	})
	if err != nil {
		return fmt.Errorf("failed to find user repository: %w", err)
	}

	var (
		title       = providers.DefaultTitle
		base        = providers.DefaultBaseBranch
		description = providers.DefaultDescription
	)

	if spec.Title != "" {
		title = spec.Title
	}

	if spec.Base != "" {
		base = spec.Base
	}

	if spec.Description != "" {
		description = spec.Description
	}

	pr, err := repo.PullRequests().Create(ctx, title, branch, base, description)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("created pull request for user repository", "user", repository.Owner, "pull-request", pr.Get().Number)

	return nil
}

// Repositories groups together a common functionality of both repository types.
type Repositories interface {
	Commits() gitprovider.CommitClient
}

func createCodeownersFile(ctx context.Context, repo Repositories, maintainers []string) error {
	if len(maintainers) == 0 {
		return nil
	}

	logger := log.FromContext(ctx)

	content := strings.Builder{}

	for _, m := range maintainers {
		_, _ = content.WriteString(fmt.Sprintf("%s\n", m))
	}

	files := []gitprovider.CommitFile{
		{
			Path:    gitprovider.StringVar("CODEOWNERS"),
			Content: gitprovider.StringVar(content.String()),
		},
	}

	commit, err := repo.Commits().Create(ctx, "main", "adding CODEOWNERS", files)
	if err != nil {
		return fmt.Errorf("failed to create CODEOWNERS file: %w", err)
	}

	logger.Info("successfully added CODEOWNERS", "url", commit.Get().URL)

	return nil
}
