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

	"github.com/containers/image/v5/pkg/compression"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-logr/logr"
	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"

	"github.com/open-component-model/ocm-controller/pkg/cache"
	"github.com/open-component-model/ocm-controller/pkg/ocm"

	"github.com/open-component-model/git-controller/pkg"
)

type Git struct {
	Logger   logr.Logger
	OciCache cache.Cache
}

func NewGoGit(log logr.Logger, cache cache.Cache) *Git {
	return &Git{
		Logger:   log,
		OciCache: cache,
	}
}

func (g *Git) Push(ctx context.Context, opts *pkg.PushOptions) (string, error) {
	g.Logger.V(v1alpha1.LevelDebug).Info(
		"running push operation",
		"msg",
		opts.Message,
		"snapshot",
		opts.Snapshot.Name,
		"url",
		opts.URL,
		"sub-path",
		opts.SubPath,
	)

	dir, err := os.MkdirTemp("", "clone")
	if err != nil {
		return "", fmt.Errorf("failed to initialize temp folder: %w", err)
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
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", opts.BaseBranch)),
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

	if opts.TargetBranch != opts.BaseBranch {
		if err := w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(opts.TargetBranch),
			Create: true,
		}); err != nil {
			return "", fmt.Errorf("failed to checkout new branch: %w", err)
		}
	}

	dir = filepath.Join(dir, opts.SubPath)
	const perm = 0o777
	if err := os.MkdirAll(dir, perm); err != nil {
		return "", fmt.Errorf("failed to create subPath: %w", err)
	}

	name, err := ocm.ConstructRepositoryName(opts.Snapshot.Spec.Identity)
	if err != nil {
		return "", fmt.Errorf("failed to construct name: %w", err)
	}

	blob, err := g.OciCache.FetchDataByDigest(ctx, name, opts.Snapshot.Spec.Digest)
	if err != nil {
		return "", fmt.Errorf("failed to fetch blob for digest: %w", err)
	}

	uncompressed, _, err := compression.AutoDecompress(blob)
	if err != nil {
		return "", fmt.Errorf("failed to auto decompress: %w", err)
	}
	defer uncompressed.Close()

	// we only care about the error if it is NOT a header error. Otherwise, we assume the content
	// wasn't compressed.
	if err = Untar(uncompressed, dir); err != nil {
		return "", fmt.Errorf("failed to untar content: %w", err)
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
	g.Logger.V(v1alpha1.LevelDebug).Info("pushing commit", "commit", commit)
	pushOptions := &git.PushOptions{
		Prune: opts.Prune,
		Auth:  auth,
	}
	if err := r.Push(pushOptions); err != nil {
		return "", fmt.Errorf("failed to push new snapshot: %w", err)
	}

	return opts.Snapshot.Spec.Digest, nil
}
