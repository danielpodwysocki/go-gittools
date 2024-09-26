package internal

import (
	"log"
	"os"
	"testing"

	"github.com/danielpodwysocki/go-gittools/internal/repository"
	"github.com/go-git/go-git/v5"
)

// initialise a parent repo, fork it and add unrelated commits to both child and parent
// This is the "happy path" where we can recommend a merge-on-green
func prepareForkedReposUnrelatedCommits() (child_repo_path, parent_repo_path string) {
	temp_path, err := os.MkdirTemp("/tmp", "go-gittools-tests")
	parent_repo_path = temp_path + "/parent_repo"
	child_repo_path = temp_path + "/child_repo"
	if err != nil {
		log.Fatalf("Failed to prep temp dir for tests: %v", err)
	}
	parent_repo, err := git.PlainInit(parent_repo_path, false)

	if err != nil {
		log.Fatalf("Failed to init test repo: %v", err)
	}

	data := []byte("Hello there! I am the first commit on this repo.")
	err = os.WriteFile(parent_repo_path+"/README.md", data, 0644)
	if err != nil {
		log.Fatalf("Failed to create README.md in the parent repo: %v", err)
	}
	// trust me.
	parent_repo_worktree, _ := parent_repo.Worktree()
	err = parent_repo_worktree.AddGlob("README.md")
	if err != nil {
		log.Fatalf("Failed to add README.md in the parent repo: %v", err)
	}
	_, err = parent_repo_worktree.Commit("add README.md", &git.CommitOptions{})
	if err != nil {
		log.Fatalf("Failed to perform the initial commit on the parent repo: %v", err)
	}

	child_repo, err := git.PlainClone(child_repo_path, false,
		&git.CloneOptions{
			URL: parent_repo_path,
		})

	if err != nil {
		log.Fatalf("Failed to clone parent repo into child: %v", err)
	}

	data = []byte("print('hello! I am pretending to be code!')")
	err = os.WriteFile(child_repo_path+"/hello.py", data, 0644)
	if err != nil {
		log.Fatalf("Failed to create hello.py in the child repo: %v", err)
	}

	// trust me.
	child_repo_worktree, _ := child_repo.Worktree()

	err = child_repo_worktree.AddGlob("hello.py")

	if err != nil {
		log.Fatalf("Failed to add hello.py in the child repo: %v", err)
	}
	_, err = child_repo_worktree.Commit("add hello.py", &git.CommitOptions{})
	if err != nil {
		log.Fatalf("Failed to commit hello.py in the child repo: %v", err)
	}

	// Diverging them, so that we can test automerging
	data = []byte("Hello there! I am the second commit on this repo.")

	err = os.WriteFile(parent_repo_path+"/README.md", data, 0644)
	if err != nil {
		log.Fatalf("Failed to overwrite README.md in the parent repo: %v", err)
	}
	err = parent_repo_worktree.AddGlob("README.md")

	if err != nil {
		log.Fatalf("Failed to add README.md in the parent repo: %v", err)
	}
	_, err = parent_repo_worktree.Commit("change README.md to test the automerge workflow", &git.CommitOptions{})
	if err != nil {
		log.Fatalf("Failed to perform the second commit on the parent repo: %v", err)
	}
	return child_repo_path, parent_repo_path
}

// ToDo: Implement&cover conflicting histories as well

func TestAutoMerge(t *testing.T) {
	child_repo_path, parent_repo_path := prepareForkedReposUnrelatedCommits()

	gitHostRepo := repository.MockGitHostRepo{}

	err := AutoMerge("file://"+child_repo_path, "file://"+parent_repo_path, "master", "master", gitHostRepo)
	if err != nil {
		log.Fatalf("AutoMerge failed with: %v", err)
	}
}
