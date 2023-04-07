// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"context"
	"fmt"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

type Provider struct {
	CreateRepositoryErr        error
	CreatePullRequestErr       error
	CreateRepositoryCalledWith map[int][]any
	CreateRepositoryCallCount  int
}

func (p *Provider) CreateRepository(ctx context.Context, obj mpasv1alpha1.Repository) error {
	if p.CreateRepositoryCalledWith == nil {
		p.CreateRepositoryCalledWith = make(map[int][]any)
	}
	p.CreateRepositoryCalledWith[p.CreateRepositoryCallCount] = append(p.CreateRepositoryCalledWith[p.CreateRepositoryCallCount], []any{obj})

	return p.CreateRepositoryErr
}

func (p *Provider) CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error {
	return p.CreatePullRequestErr
}

func (p *Provider) CreateRepositoryCallArgsForNumber(i int) ([]any, error) {
	args, ok := p.CreateRepositoryCalledWith[i]
	if !ok {
		return nil, fmt.Errorf("arguments for cal number %d not found", i)
	}

	return args, nil
}

func NewProvider() *Provider {
	return &Provider{}
}

var _ providers.Provider = &Provider{}
