// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gogit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-logr/logr"

	"github.com/open-component-model/git-sync-controller/pkg"
)

type Git struct {
	Logger    logr.Logger
	OciClient pkg.OCIClient
}

func NewGoGit(log logr.Logger, ociClient pkg.OCIClient) *Git {
	return &Git{
		Logger:    log,
		OciClient: ociClient,
	}
}

func (g *Git) Push(ctx context.Context, opts *pkg.PushOptions) (string, error) {
	g.Logger.V(4).Info("running push operation", "msg", opts.Message, "snapshot", opts.SnapshotURL, "url", opts.URL, "sub-path", opts.SubPath)

	dir, err := os.MkdirTemp("", "clone")
	if err != nil {
		return "", fmt.Errorf("failed to initialise temp folder: %w", err)
	}

	var auth transport.AuthMethod
	if opts.Auth != nil {
		if v := opts.Auth.BasicAuth; v != nil {
			auth = &http.BasicAuth{
				Username: v.Username,
				Password: v.Password,
			}
		}
		if v := opts.Auth.SSH; v != nil {
			pb, err := ssh.NewPublicKeys(v.User, v.PemBytes, v.Password)
			if err != nil {
				return "", fmt.Errorf("failed to create public key authentication: %w", err)
			}
			auth = pb
		}
	}

	cloneOptions := &git.CloneOptions{
		URL:           opts.URL,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", opts.Branch)),
		Auth:          auth,
	}

	r, err := git.PlainClone(dir, false, cloneOptions)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to create a worktree: %w", err)
	}

	dir = filepath.Join(dir, opts.SubPath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", fmt.Errorf("failed to create subPath: %w", err)
	}

	//Pull will result in an untar-ed list of files.

	digest, err := g.OciClient.Pull(ctx, opts.SnapshotURL, dir)
	if err != nil {
		return "", fmt.Errorf("failed to pull from OCI repository: %w", err)
	}

	// Add all extracted files.
	if err := w.AddGlob("."); err != nil {
		return "", fmt.Errorf("failed to add items to worktree: %w", err)
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
		return "", fmt.Errorf("failed to commit changes: %w", err)
	}
	g.Logger.V(4).Info("pushing commit", "commit", commit)
	pushOptions := &git.PushOptions{
		Prune: opts.Prune,
		Auth:  auth,
	}
	if err := r.Push(pushOptions); err != nil {
		return "", fmt.Errorf("failed to push new snapshot: %w", err)
	}
	return digest, nil
}
