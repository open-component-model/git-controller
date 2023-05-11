// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gogit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CreateOrganizationRepository creates a repository for an authenticated organization.
func CreateOrganizationRepository(ctx context.Context, gc gitprovider.Client, domain string, obj mpasv1alpha1.Repository) error {
	logger := log.FromContext(ctx)

	visibility := gitprovider.RepositoryVisibility(obj.Spec.Visibility)

	if err := gitprovider.ValidateRepositoryVisibility(visibility); err != nil {
		return fmt.Errorf("failed to validate visibility: %w", err)
	}

	ref := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: obj.Spec.Owner,
		},
		RepositoryName: obj.GetName(),
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar(obj.Spec.DefaultBranch),
		Visibility:    &visibility,
	}

	createOpts, err := gitprovider.MakeRepositoryCreateOptions(&gitprovider.RepositoryCreateOptions{AutoInit: gitprovider.BoolVar(true)})
	if err != nil {
		return fmt.Errorf("failed to create _create_ options for repository: %w", err)
	}

	switch obj.Spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		repo, err := gc.OrgRepositories().Create(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		if err := setupProjectStructure(ctx, repo, obj.Spec.Maintainers); err != nil {
			if cerr := repo.Delete(ctx); cerr != nil {
				err = errors.Join(err, cerr)
			}

			return fmt.Errorf("failed to create initial project structure: %w", err)
		}

		logger.Info("successfully created organization repository", "domain", domain, "repository", obj.GetName())
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		repo, created, err := gc.OrgRepositories().Reconcile(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", obj.GetName())
		} else {
			if err := setupProjectStructure(ctx, repo, obj.Spec.Maintainers); err != nil {
				if cerr := repo.Delete(ctx); cerr != nil {
					err = errors.Join(err, cerr)
				}

				return fmt.Errorf("failed to create initial project structure: %w", err)
			}

			logger.Info("successfully created organization repository", "domain", domain, "repository", obj.GetName())
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", obj.Spec.ExistingRepositoryPolicy)
	}

	return nil
}

// CreateUserRepository creates a repository for an authenticated user.
func CreateUserRepository(ctx context.Context, gc gitprovider.Client, domain string, obj mpasv1alpha1.Repository) error {
	logger := log.FromContext(ctx)

	visibility := gitprovider.RepositoryVisibility(obj.Spec.Visibility)

	if err := gitprovider.ValidateRepositoryVisibility(visibility); err != nil {
		return fmt.Errorf("failed to validate visibility: %w", err)
	}

	ref := gitprovider.UserRepositoryRef{
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: obj.Spec.Owner,
		},
		RepositoryName: obj.GetName(),
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("main"),
		Visibility:    &visibility,
	}

	createOpts, err := gitprovider.MakeRepositoryCreateOptions(&gitprovider.RepositoryCreateOptions{AutoInit: gitprovider.BoolVar(true)})
	if err != nil {
		return fmt.Errorf("failed to create _create_ options for repository: %w", err)
	}

	switch obj.Spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		repo, err := gc.UserRepositories().Create(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		if err := setupProjectStructure(ctx, repo, obj.Spec.Maintainers); err != nil {
			if cerr := repo.Delete(ctx); cerr != nil {
				err = errors.Join(err, cerr)
			}

			return fmt.Errorf("failed to create initial project structure: %w", err)
		}

		logger.Info("successfully created user repository", "domain", domain, "repository", obj.GetName())
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		repo, created, err := gc.UserRepositories().Reconcile(ctx, ref, info, &createOpts)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", obj.GetName())
		} else {
			if err := setupProjectStructure(ctx, repo, obj.Spec.Maintainers); err != nil {
				if cerr := repo.Delete(ctx); cerr != nil {
					err = errors.Join(err, cerr)
				}

				return fmt.Errorf("failed to create initial project structure: %w", err)
			}

			logger.Info("successfully created user repository", "domain", domain, "repository", obj.GetName())
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", obj.Spec.ExistingRepositoryPolicy)
	}

	return nil
}

// CreateOrganizationPullRequest creates a pull-request for an organization owned repository.
func CreateOrganizationPullRequest(ctx context.Context, gc gitprovider.Client, domain, branch string, spec deliveryv1alpha1.PullRequestTemplate, repository mpasv1alpha1.Repository) error {
	// find the repository
	repo, err := gc.OrgRepositories().Get(ctx, gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: repository.Spec.Owner,
		},
		RepositoryName: repository.GetName(),
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
	logger.Info("created pull request for organization repository", "organization", repository.Spec.Owner, "pull-request", pr.Get().Number)

	return nil
}

// CreateUserPullRequest creates a pull-request for a user owned repository.
func CreateUserPullRequest(ctx context.Context, gc gitprovider.Client, domain, branch string, spec deliveryv1alpha1.PullRequestTemplate, repository mpasv1alpha1.Repository) error {
	// find the repository
	repo, err := gc.UserRepositories().Get(ctx, gitprovider.UserRepositoryRef{
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: repository.Spec.Owner,
		},
		RepositoryName: repository.GetName(),
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
	logger.Info("created pull request for user repository", "user", repository.Spec.Owner, "pull-request", pr.Get().Number)

	return nil
}

// Repositories groups together a common functionality of both repository types.
type Repositories interface {
	Commits() gitprovider.CommitClient
}

func setupProjectStructure(ctx context.Context, repo Repositories, maintainers []string) error {
	logger := log.FromContext(ctx)

	var files []gitprovider.CommitFile

	if len(maintainers) > 0 {
		content := strings.Builder{}

		for _, m := range maintainers {
			_, _ = content.WriteString(fmt.Sprintf("%s\n", m))
		}

		files = append(files, gitprovider.CommitFile{
			Path:    gitprovider.StringVar("CODEOWNERS"),
			Content: gitprovider.StringVar(content.String()),
		})
	}

	files = append(files, gitprovider.CommitFile{
		Path:    gitprovider.StringVar("generators/.keep"),
		Content: gitprovider.StringVar(""),
	}, gitprovider.CommitFile{
		Path:    gitprovider.StringVar("products/.keep"),
		Content: gitprovider.StringVar(""),
	}, gitprovider.CommitFile{
		Path:    gitprovider.StringVar("subscriptions/.keep"),
		Content: gitprovider.StringVar(""),
	}, gitprovider.CommitFile{
		Path:    gitprovider.StringVar("targets/.keep"),
		Content: gitprovider.StringVar(""),
	})

	commit, err := repo.Commits().Create(ctx, "main", "creating initial project structure", files)
	if err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	logger.Info("successfully created initial project structure", "url", commit.Get().URL)

	return nil
}
