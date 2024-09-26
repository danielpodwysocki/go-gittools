package main

import (
	"github.com/danielpodwysocki/go-gittools/internal"
	"github.com/danielpodwysocki/go-gittools/internal/repository"
)

func main() {
	// cmd.Parse()
	// hardcoded until we suppopt more
	providerName := "github"

	var provider repository.GitHostRepo

	switch providerName {
	case "github":
		provider = repository.GitHubGitHostRepo{}
	}

	internal.AutoMerge("https://github.com/git-fixtures/basic.git", "https://github.com/git-fixtures/basic.git", "master", "master", provider)
}
