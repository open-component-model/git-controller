package v1alpha1

const (
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
)
