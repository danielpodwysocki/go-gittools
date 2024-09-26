package repository

import "fmt"

// gnarly name.
type GitHubGitHostRepo struct {
	username string
}

func (r GitHubGitHostRepo) TriggerMergeOnGreen(repoURL string, sourceBranch string, targetBranch string) error {
	return fmt.Errorf("Finish implementing TriggerMergeOnGreen!")

}
