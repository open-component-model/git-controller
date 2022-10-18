package gogit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-logr/logr"
)

type Git struct {
	Logger logr.Logger
}

func NewGit(log logr.Logger) *Git {
	return &Git{
		Logger: log,
	}
}

func (g *Git) Push(ctx context.Context, msg, snapshotLocation, destination string) error {
	g.Logger.V(4).Info("running push operation", "msg", msg, "snapshot", snapshotLocation, "destination", destination)
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
		URLs: []string{destination},
	}); err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}
	// TODO: Figure out if this is needed at all.
	// Probably yes, since we are pushing into a gitops repository.
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec("refs/heads/main")},
		Depth:    1,
	}
	// Detect what kind of Auth is provided.
	fetchOptions.Auth = &http.BasicAuth{
		Username: "Skarlso",
		Password: "token",
	}
	if err := r.Fetch(fetchOptions); err != nil {
		return fmt.Errorf("failed to fetch remote ref '%s': %w", "main", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create a worktree: %w", err)
	}

	// TODO: move the snapshot now
	// Extract or add as is?
	if err := w.AddGlob("."); err != nil {
		return fmt.Errorf("failed to add items to worktree: %w", err)
	}
	opts := git.CommitOptions{
		Author: &object.Signature{
			Name:  "Gergely Brautigam",
			Email: "182850+Skarlso@users.noreply.github.com",
			When:  time.Now(),
		},
	}
	commit, err := w.Commit("Uploading snapshot to location", &opts)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	g.Logger.V(4).Info("pushing commit", "commit", commit)
	if err := r.Push(&git.PushOptions{}); err != nil {
		return fmt.Errorf("failed to push new snapshot: %w", err)
	}
	return nil
}
