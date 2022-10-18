package oci

import (
	"context"
	"fmt"

	"github.com/fluxcd/pkg/oci/client"
	"github.com/google/go-containerregistry/pkg/crane"
)

type Client struct {
	url    string
	client *client.Client
}

// NewClient creates a new OCI client with target URL and user agent.
func NewClient(url, agent string) *Client {
	options := []crane.Option{
		crane.WithUserAgent(agent),
	}
	client := client.NewClient(options)

	return &Client{
		url:    url,
		client: client,
	}
}

// Pull takes a snapshot name and pulls it from the OCI repository.
func (o *Client) Pull(ctx context.Context, url, outDir string) error {
	if _, err := o.client.Pull(ctx, url, outDir); err != nil {
		return fmt.Errorf("failed to pull snapshot: %w", err)
	}

	return nil
}
