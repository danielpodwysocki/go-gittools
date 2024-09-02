package porcelain

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// Checkout and branch creation logic is a little muddy
// git.Repository.CreateBranch simply puts a branch spec inside .git/config, however it DOES not create the actual ref
// That makes it necessary to 2-step this - this file provides some pretty porcelain around

// Here be dragons:
// https://github.com/go-git/go-git/blob/master/_examples/branch/main.go
// https://github.com/go-git/go-git/issues/632

// checkout a branch, create and checkout if not present.
func CheckoutOrCreateBranch(repo git.Repository, branchName string) error {
	return nil
}

//func BranchExists(repo git.Repository, branch_name string) (exists bool, err error) {
//	branches, err := repo.Branches()
//	if err != nil {
//		return false, err
//	}
//
//}

// Create a branch and check it out. Return an error if it already exists!
func CreateAndCheckoutBranch(repo *git.Repository, branch_name string, branch_from plumbing.Hash) error {
	err := CreateBranch(repo, branch_name, branch_from)
	if err != nil {
		return err
	}
	err = CheckoutBranch(repo, plumbing.NewBranchReferenceName(branch_name))
	if err != nil {
		return err
	}
	return nil
}

func CreateBranch(repo *git.Repository, branch_name string, branch_from plumbing.Hash) error {
	branchRefName := plumbing.NewBranchReferenceName(branch_name)
	branchRef := plumbing.NewHashReference(branchRefName, branch_from)
	err := repo.Storer.SetReference(branchRef)
	if err != nil {
		return err
	}
	repo.CreateBranch(&config.Branch{
		Name:   branch_name,
		Remote: "origin",
	})
	fmt.Println("Created branch", branchRefName, "pointing at", branch_from)
	return nil
}

// Get the hash of a commit at the head of a branch
func GetBranchHash(repo *git.Repository, branch string) (hash plumbing.Hash, err error) {
	refName := plumbing.ReferenceName(branch)
	ref, err := repo.Reference(refName, true)
	if err != nil {
		return *new(plumbing.Hash), fmt.Errorf("failed to find branch ref %v: %v", branch, err)
	}

	return ref.Hash(), nil
}

func CheckoutBranch(repo *git.Repository, branchRef plumbing.ReferenceName) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	err = worktree.Checkout(
		&git.CheckoutOptions{
			Branch: branchRef,
		})
	if err != nil {
		return fmt.Errorf("failed to get checkout %v: %v", branchRef, err)
	}

	return nil
}
