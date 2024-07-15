package internal

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

func AutoMerge(fork_repo string, origin_repo string) {
	clone_root := "/tmp/foo"
	prepRepo(fork_repo, clone_root)
	prepRepo(origin_repo, clone_root)

}

func prepRepo(repo_url string, clone_root string) {
	// Clones down the repo, resets it to the origin/main or origin/master branch
	_, err := git.PlainClone(clone_root, false,
		&git.CloneOptions{
			URL: repo_url,
		})
	if err != nil {
		fmt.Println("Whatup")
	}
}
