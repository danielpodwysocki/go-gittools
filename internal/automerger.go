package internal

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func getRepoNameFromURL(repo_url string) string {

	repo_url_split := strings.Split(repo_url, "/")
	repo_name := strings.Trim(repo_url_split[len(repo_url_split)-1], ".git")
	return repo_name
}

func AutoMerge(forked_repo_url string, parent_origin_repo_url string) error {
	clone_root := "/tmp/foo/"

	err := prepRepo(forked_repo_url, clone_root+getRepoNameFromURL(forked_repo_url))
	if err != nil {
		return fmt.Errorf("failed to prep forked/child repo: %v", err)
	}
	err = prepRepo(parent_origin_repo_url, clone_root+getRepoNameFromURL((parent_origin_repo_url)))
	if err != nil {
		return fmt.Errorf("failed to prep parent/origin repo: %v", err)
	}

	forked_repo, err := git.PlainOpen(clone_root + getRepoNameFromURL(forked_repo_url))
	if err != nil {
		return fmt.Errorf("failed to open forked/child repo: %v", err)
	}
	err = ensureRemote(forked_repo, "parent_origin", parent_origin_repo_url)
	if err != nil {
		return fmt.Errorf("failed to add the parent repo as an origin on the child/fork: %v", err)
	}
	today_iso8601 := time.Now().Format(time.RFC3339)
	//today_iso8601 = strings.Replace(":", "-", today_iso8601, -1)
	branch_name := "go-gittools-automerge-" + today_iso8601
	err = forked_repo.CreateBranch(&config.Branch{
		Name: "testujemy",
	})
	if err != nil {
		return fmt.Errorf("failed to create branch %v: %v", branch_name, err)
	}
	return nil

}

func ensureRemote(repo *git.Repository, remote_name string, remote_url string) error {
	// Adds a remote if it doesn't exist
	remotes, err := repo.Remotes()
	if err != nil {
		return errors.New("failed to get remotes: " + err.Error())
	}
	for _, remote := range remotes {
		if remote.Config().Name == remote_name {
			return nil
		}
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: remote_name,
		URLs: []string{remote_url},
	})

	if err != nil {
		return errors.New("failed to create remote: " + err.Error())
	}
	log.Printf("Added remote %v with url %v", remote_name, remote_url)
	return nil
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
