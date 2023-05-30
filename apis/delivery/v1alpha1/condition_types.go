// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

const (
	// PatchFailedReason is used when we couldn't patch an object.
	PatchFailedReason = "PatchFailed"

	// SnapshotGetFailedReason is used when the needed snapshot does not exist.
	SnapshotGetFailedReason = "SnapshotGetFailed"

	// RepositoryGetFailedReason is used when the needed repository does not exist.
	RepositoryGetFailedReason = "RepositoryGetFailed"

	// CredentialsNotFoundReason is used when the needed authentication does not exist.
	CredentialsNotFoundReason = "CredentialsNotFound"

	// GitRepositoryPushFailedReason is used when the needed pushing to a git repository failed.
	GitRepositoryPushFailedReason = "GitRepositoryPushFailed"

	// CreatePullRequestFailedReason is used when creating a pull request failed.
	CreatePullRequestFailedReason = "CreatePullRequestFailed"

	// GitRepositoryCreateFailedReason is used when creating a git repository failed.
	GitRepositoryCreateFailedReason = "GitRepositoryCreateFailed"
)
