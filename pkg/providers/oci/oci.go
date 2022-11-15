// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"

	"github.com/fluxcd/pkg/oci/client"
	"github.com/google/go-containerregistry/pkg/crane"
)

type Client struct {
	client *client.Client
}

// NewClient creates a new OCI client with target URL and user agent.
func NewClient(agent string) *Client {
	options := []crane.Option{
		crane.WithUserAgent(agent),
	}
	client := client.NewClient(options)

	return &Client{
		client: client,
	}
}

// Pull takes a snapshot name and pulls it from the OCI repository.
func (o *Client) Pull(ctx context.Context, url, outDir string) (string, error) {
	m, err := o.client.Pull(ctx, url, outDir)
	if err != nil {
		return "", fmt.Errorf("failed to pull snapshot: %w", err)
	}

	return m.Digest, nil
}
