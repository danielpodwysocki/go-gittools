package internal

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
)

func getRepoNameFromURL(repo_url string) string {

	repo_url_split := strings.Split(repo_url, "/")
	repo_name := strings.Trim(repo_url_split[len(repo_url_split)-1], ".git")
	return repo_name
}

func AutoMerge(fork_repo_url string, origin_repo_url string) {
	clone_root := "/tmp/foo/"

	err := prepRepo(fork_repo_url, clone_root+getRepoNameFromURL(fork_repo_url))
	if err != nil {
		panic(err)
	}
	err = prepRepo(origin_repo_url, clone_root+getRepoNameFromURL((origin_repo_url)))
	if err != nil {
		panic(err)
	}

}

func prepRepo(repo_url string, clone_path string) error {
	// Clones down the repo, resets it to the origin/main or origin/master branch
	repo, err := git.PlainOpen(clone_path)
	if err == nil {
		// a little ugly - ideally would sit in the resetRepo function, see comments inside it
		log.Printf("Resetting %v ,located at %v", repo_url, clone_path)
		err = fetchRepo(repo)
		if err != nil {
			return err
		}
		err = resetRepo(repo, "master")
		if err != nil {
			return err
		}
		err = pullCurrentBranch(repo)
		if err != nil {
			return err
		}

		return nil
	}
	_, err = git.PlainClone(clone_path, false,
		&git.CloneOptions{
			URL: repo_url,
		})
	if err != nil {
		return errors.New("failed to clone repo for prep: " + err.Error())
	}

	return nil
}

func fetchRepo(repo *git.Repository) error {
	// Fetches the repo
	err := repo.Fetch(&git.FetchOptions{})
	log.Printf("Fetching repo")
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return errors.New("failed to fetch repo: " + err.Error())
	}
	return nil
}

func checkoutBranch(repo *git.Repository, branch_name string) error {
	// Checks out the branch
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.New("failed to prep worktree: " + err.Error())
	}
	//branch = fmt.Sprintf("refs/remotes/origin/%v", branch)
	branch, err := repo.Branch(branch_name)
	if err != nil {
		return fmt.Errorf("failed to get branch %v : %v", branch_name, err)
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		// branch.Merge is the refspec
		Branch: branch.Merge,
		Force:  true,
	})

	if err != nil {
		return errors.New("failed to checkout branch: " + err.Error())
	}
	return nil
}

// https://github.com/src-d/go-git/issues/559
// This needs to stay slightly unaware of paths (or get them passed explicitly if needed in the future)
// The only way to derive them from git.Repository is not guaranteed to continue working.
func resetRepo(repo *git.Repository, branch string) error {

	worktree, err := repo.Worktree()

	if err != nil {
		return errors.New("failed to prep worktree: " + err.Error())
	}
	_, err = repo.Branch(branch)
	if err != nil {
		return fmt.Errorf("failed to get branch %v : %v", branch, err)
	}
	err = worktree.Reset(
		&git.ResetOptions{
			Mode: git.HardReset,
		})
	if err != nil {
		return errors.New("failed to reset repo: " + err.Error())
	}
	err = checkoutBranch(repo, branch)
	if err != nil {
		return err
	}

	return nil
}

func pullCurrentBranch(repo *git.Repository) error {
	log.Printf("Pulling current branch")
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.New("failed to prep worktree: " + err.Error())
	}
	err = worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return errors.New("failed to pull repo: " + err.Error())
	}
	return nil
}
