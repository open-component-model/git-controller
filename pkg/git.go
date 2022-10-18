package pkg

import "context"

// Git defines an interface to abstract git operations.
type Git interface {
	Push(ctx context.Context, msg, location, destination string) error
}
