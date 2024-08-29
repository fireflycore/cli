package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetRepoVersion(repo string, version string) {
	if version == "" {
		version = "latest"

		// todo 获取最新
	}

	// 先从本地查看有无该版本号的缓存
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	templateCacheDir := fmt.Sprintf("%s/%s/template/%s", cacheDir, CliName, version)
	info, err := os.Stat(templateCacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			// todo 文件夹不存在, 获取最新版本
		} else {
			panic(err)
		}
	} else if !info.IsDir() {
		// todo 文件夹不存在，获取最新版本
	} else {
		// todo 文件夹存在，将当前版本的模版copy一份-> template/temp, 项目创建完成后将 temp 目录清空
	}

	url := fmt.Sprintf("https://api.github.com/repos/lhdhtrc/microservice-go/releases/%s", version)
	fmt.Println(url)
}

func GetRepo() {
	var version string
	GetRepoVersion("", version)

	repo := "https://github.com/lhdhtrc/microservice-go"
	language := "Go"

	cacheDir := fmt.Sprintf("./.firefly/cache/template/%s/%", strings.ToLower(language), "")

	cmd := exec.Command("git", "clone", "--depth=1", repo, cacheDir)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
