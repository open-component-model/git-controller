package gogit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-logr/logr"

	"github.com/open-component-model/git-sync-controller/pkg"
)

type Git struct {
	Logger logr.Logger
	Client pkg.OCIClient
}

func NewGoGit(log logr.Logger, ociClient pkg.OCIClient) *Git {
	return &Git{
		Logger: log,
		Client: ociClient,
	}
}

func (g *Git) Push(ctx context.Context, opts *pkg.PushOptions) error {
	g.Logger.V(4).Info("running push operation", "msg", opts.Message, "snapshot", opts.SnapshotURL, "url", opts.URL, "sub-path", opts.SubPath)
	// Get the snapshot from snapshotLocation
	// move to this tmp folder or ( Fetch ) to this tmp folder once the git remote is initialised?
	dir, err := os.MkdirTemp("", "clone")
	if err != nil {
		return fmt.Errorf("failed to initialise temp folder: %w", err)
	}

	fs := osfs.New(dir)
	r, err := git.Init(filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil)
	if err != nil {
		return fmt.Errorf("failed to initialise git repo: %w", err)
	}

	if _, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{opts.URL},
	}); err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}

	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(opts.Branch)},
		Depth:    1,
	}

	if opts.Auth != nil {
		if v := opts.Auth.BasicAuth; v != nil {
			fetchOptions.Auth = &http.BasicAuth{
				Username: v.Username,
				Password: v.Password,
			}
		}
		if v := opts.Auth.SSH; v != nil {
			pb, err := ssh.NewPublicKeys(v.User, v.PemBytes, v.Password)
			if err != nil {
				return fmt.Errorf("failed to create public key authentication: %w", err)
			}
			fetchOptions.Auth = pb
		}
	}

	if err := r.Fetch(fetchOptions); err != nil {
		return fmt.Errorf("failed to fetch remote ref '%s': %w", "main", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create a worktree: %w", err)
	}

	dir = filepath.Join(dir, opts.SubPath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("failed to create subPath: %w", err)
	}
	// Pull will result in an untar-ed list of files.
	if err := g.Client.Pull(ctx, opts.SnapshotURL, dir); err != nil {
		return fmt.Errorf("failed to pull from OCI repository: %w", err)
	}
	// Add all extracted files.
	if err := w.AddGlob("."); err != nil {
		return fmt.Errorf("failed to add items to worktree: %w", err)
	}

	commitOpts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  opts.Name,
			Email: opts.Email,
			When:  time.Now(),
		},
	}
	commit, err := w.Commit("Uploading snapshot to location", commitOpts)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	g.Logger.V(4).Info("pushing commit", "commit", commit)
	if err := r.Push(&git.PushOptions{}); err != nil {
		return fmt.Errorf("failed to push new snapshot: %w", err)
	}
	return nil
}
