// Copyright 2022.
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

const (
	// PatchFailedReason is used when we couldn't patch an object.
	PatchFailedReason = "PatchFailed"

	// SnapshotGetFailedReason is used when the needed snapshot does not exist.
	SnapshotGetFailedReason = "SnapshotGetFailed"

	// AuthenticateGetFailedReason is used when the needed authentication does not exist.
	AuthenticateGetFailedReason = "AuthenticateGetFailed"

	// GitRepositoryPushFailedReason is used when the needed pushing to a git repository failed.
	GitRepositoryPushFailedReason = "GitRepositoryPushFailed"
)
