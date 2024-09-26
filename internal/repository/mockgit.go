package repository

type MockGitHostRepo struct {
	username string
}

func (r MockGitHostRepo) TriggerMergeOnGreen(repoURL string, sourceBranch string, targetBranch string) error {
	return nil
}
