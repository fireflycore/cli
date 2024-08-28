package create

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetRepo() {
	repo := "https://github.com/lhdhtrc/microservice-go"
	language := "Go"

	cacheDir := fmt.Sprintf("./.firefly/cache/template/%s/%", strings.ToLower(language), "")

	cmd := exec.Command("git", "clone", "--depth=1", repo, cacheDir)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
