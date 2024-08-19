package internal

import (
	"log"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
)

// initialise a parent repo, fork it and add unrelated commits to both child and parent
// This is the "happy path" where we can recommend a merge-on-green
func prepareForkedReposUnrelatedCommits() {
	temp_path, err := os.MkdirTemp("/tmp", "parent_repo")
	if err != nil {
		log.Fatalf("Failed to prep temp dir for tests: %v", err)
	}
	repo, err := git.PlainInit(temp_path, false)

	if err != nil {
		log.Fatalf("Failed to init test repo: %v", err)
	}

	data := []byte("Hello there! I am the first commit on this repo.")
	err = os.WriteFile(temp_path+"/README.md", data, 0644)
	if err != nil {
		log.Fatalf("Failed to create README.md in the parent repo: %v", err)
	}
	repo.Branch("master")
}

// ToDo: Cover conflicting histories as well

func TestAutoMerge(t *testing.T) {
	forked_repo_url := "https://test.test123xyz"
	parent_origin_repo_url := "https://test.test123xyz"

	prepareForkedReposUnrelatedCommits()

	err := AutoMerge(forked_repo_url, parent_origin_repo_url)
	if err != nil {
		log.Fatalf("AutoMerge failed with: %v", err)
	}
}
