package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/danielpodwysocki/go-gittools/internal/porcelain"
	"github.com/danielpodwysocki/go-gittools/internal/repository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	gogsgit "github.com/gogs/git-module"
)

func getRepoNameFromURL(repo_url string) string {

	repo_url_split := strings.Split(repo_url, "/")
	repo_name := strings.Trim(repo_url_split[len(repo_url_split)-1], ".git")
	return repo_name
}

// todo: finish implementation
func AutoMerge(child_repo_url string, parent_repo_url string, child_repo_branch string, parent_repo_branch string, gitHostRepo repository.GitHostRepo) error {
	clone_root, err := os.MkdirTemp("/tmp", "go-gittools-automerge")
	if err != nil {
		return fmt.Errorf("failed preparing a temp dir: %v", err)
	}

	child_repo_path := clone_root + "/" + getRepoNameFromURL(child_repo_url)
	parent_repo_path := clone_root + "/" + getRepoNameFromURL(parent_repo_url)

	err = prepRepo(child_repo_url, child_repo_path)
	if err != nil {
		return fmt.Errorf("failed to prep child repo: %v", err)
	}
	err = prepRepo(parent_repo_url, parent_repo_path)
	if err != nil {
		return fmt.Errorf("failed to prep parent/origin repo: %v", err)
	}

	child_repo, err := git.PlainOpen(child_repo_path)
	fmt.Println(child_repo_path)
	if err != nil {
		return fmt.Errorf("failed to open child repo: %v", err)
	}
	err = ensureRemote(child_repo, "parent_origin", parent_repo_url)
	if err != nil {
		return fmt.Errorf("failed to add the parent repo as an origin on the child/fork: %v", err)
	}

	err = child_repo.Fetch(&git.FetchOptions{
		RemoteName: "parent_origin",
	})

	if err != nil {
		return fmt.Errorf("failed to fetch the parent repo on the child/fork: %v", err)
	}

	today_iso8601 := time.Now().Format(time.RFC3339)
	automergeBranchName := "go-gittools-automerge-" + strings.Replace(today_iso8601, ":", "--", -1)
	branch_from, err := porcelain.GetBranchHash(child_repo, fmt.Sprintf("refs/remotes/origin/%v", child_repo_branch))

	if err != nil {
		return err
	}

	err = porcelain.CreateAndCheckoutBranch(child_repo, automergeBranchName, branch_from)

	if err != nil {
		return err
	}
	// Useful once go-git supports this:
	// parentRepoRefStr := fmt.Sprintf("refs/remotes/parent_origin/%v", parent_repo_branch)
	// targetBranch := plumbing.NewBranchReferenceName(parentRepoRefStr)
	// targetHash, err := porcelain.GetBranchHash(child_repo, parentRepoRefStr)
	// targetRef := plumbing.NewHashReference(targetBranch, targetHash)

	// go-git only supports a fast forward merge
	// ToDo: once the ort/3-way-merge is supported, replace this and get rid of
	// the extra dependency on real git
	mergeCommand := gogsgit.NewCommand("merge", fmt.Sprintf("parent_origin/%v", parent_repo_branch))
	_, err = mergeCommand.RunInDir(child_repo_path)

	// ToDo: verify this is clear enough and we don't actually need stdout/stderr from the above
	if err != nil {
		return err
	}

	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%v:refs/remotes/origin/%v", automergeBranchName, automergeBranchName))
	err = child_repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{refSpec},
	})
	if err != nil {
		return err
	}

	err = gitHostRepo.TriggerMergeOnGreen(child_repo_url, automergeBranchName, child_repo_branch)
	if err != nil {
		return err
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
		return fmt.Errorf("failed to clone %v repo for prep: %v", repo_url, err.Error())
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
