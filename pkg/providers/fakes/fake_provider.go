// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"context"
	"fmt"

	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

type Provider struct {
	CreateRepositoryErr         error
	CreateRepositoryCalledWith  map[int][]any
	CreateRepositoryCallCount   int
	CreatePullRequestErr        error
	CreatePullRequestCalledWith map[int][]any
	CreatePullRequestCallCount  int
}

var _ providers.Provider = &Provider{}

func (p *Provider) CreateRepository(ctx context.Context, obj mpasv1alpha1.Repository) error {
	if p.CreateRepositoryCalledWith == nil {
		p.CreateRepositoryCalledWith = make(map[int][]any)
	}
	p.CreateRepositoryCalledWith[p.CreateRepositoryCallCount] = append(p.CreateRepositoryCalledWith[p.CreateRepositoryCallCount], obj)
	p.CreateRepositoryCallCount++

	return p.CreateRepositoryErr
}

func (p *Provider) CreatePullRequest(ctx context.Context, branch string, sync deliveryv1alpha1.Sync, repository mpasv1alpha1.Repository) error {
	if p.CreatePullRequestCalledWith == nil {
		p.CreatePullRequestCalledWith = make(map[int][]any)
	}
	p.CreatePullRequestCalledWith[p.CreatePullRequestCallCount] = append(p.CreatePullRequestCalledWith[p.CreatePullRequestCallCount], branch, sync, repository)
	p.CreatePullRequestCallCount++

	return p.CreatePullRequestErr
}

func (p *Provider) CreatePullRequestCallArgsForNumber(i int) ([]any, error) {
	args, ok := p.CreatePullRequestCalledWith[i]
	if !ok {
		return nil, fmt.Errorf("arguments for cal number %d not found", i)
	}

	return args, nil
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
