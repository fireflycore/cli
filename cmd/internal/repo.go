package internal

import "fmt"

func GetRepoVersion(repo string, version string) {
	if version == "" {
		version = "latest"
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/microservice-go/releases/%s", REPO_OWNER, version)
	fmt.Println(url)
}
