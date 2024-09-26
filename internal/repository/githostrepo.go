package repository

type GitHostRepo interface {
	TriggerMergeOnGreen(repoURL string, sourceBranch string, targetBranch string) error
}
