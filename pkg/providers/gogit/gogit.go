// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gogit

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
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

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		if _, err := gc.OrgRepositories().Create(ctx, ref, info); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		logger.Info("successfully created organization repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		_, created, err := gc.OrgRepositories().Reconcile(ctx, ref, info)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
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

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		if _, err := gc.UserRepositories().Create(ctx, ref, info); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		_, created, err := gc.UserRepositories().Reconcile(ctx, ref, info)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
			logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", spec.ExistingRepositoryPolicy)
	}

	return nil
}
