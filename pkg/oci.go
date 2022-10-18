package pkg

import "context"

// OCIClient defines the needed capabilities of a client that can interact with an OCI repository.
type OCIClient interface {
	Pull(ctx context.Context, url, outDir string) error
}
