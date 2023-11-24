// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package providers

import (
	"context"
	"errors"

	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
)

const (
	DefaultTitle       = "Git Controller automated Pull Request"
	DefaultBaseBranch  = "main"
	DefaultDescription = "Pull requested created automatically by OCM Git Controller."
)

var ErrNotSupported = errors.New("functionality not supported by provider")

// Provider adds the ability to create repositories and pull requests.
type Provider interface {
	CreateRepository(ctx context.Context, obj mpasv1alpha1.Repository) error
	CreatePullRequest(ctx context.Context, branch string, sync deliveryv1alpha1.Sync, repository mpasv1alpha1.Repository) (int, error)
	CreateBranchProtection(ctx context.Context, obj mpasv1alpha1.Repository) error
}
