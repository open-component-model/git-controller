package pkg

import "context"

// BasicAuth provides information for basic authentication. The expected format is Username as username and
// Password is usually a token.
type BasicAuth struct {
	Username string
	Password string
}

// SSH provides authentication for private repositories using an SSH key. The key has to be the Private key.
type SSH struct {
	PemBytes []byte
	User     string
	Password string
}

// Auth defines authentication options for repositories.
type Auth struct {
	BasicAuth *BasicAuth
	SSH       *SSH
}

// PushOptions contains settings for a push action.
type PushOptions struct {
	Auth        *Auth
	URL         string
	Message     string
	Name        string
	Email       string
	SnapshotURL string
	Branch      string
	SubPath     string
	Prune       bool
}

// Git defines an interface to abstract git operations.
type Git interface {
	Push(ctx context.Context, opts *PushOptions) (string, error)
}
