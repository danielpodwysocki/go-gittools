package main

import (
	"github.com/danielpodwysocki/go-gittools/internal"
)

func main() {
	// cmd.Parse()
	internal.AutoMerge("https://github.com/git-fixtures/basic.git", "https://github.com/git-fixtures/basic.git")
}
