package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func GetTemplateRepo(language string) string {
	var template string

	switch language {
	case "Go":
		template = "microservice-go"
	case "NodeJS":
		template = "microservice-node"
	case "Rust":
		template = "microservice-rust"
	}

	return template
}

func GetRepoUrl(language string) string {
	return fmt.Sprintf("https://github.com/lhdhtrc/%s.git", GetTemplateRepo(language))
}

func GetRepoLocal(language string, version string, dir string) error {
	cmd := exec.Command("git", "clone", GetRepoUrl(language), dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("get repo clone: %s", err)
	}

	cmdCheckout := exec.Command("git", "checkout", version, "--force")
	cmdCheckout.Dir = dir // 设置工作目录为已克隆的仓库
	if err := cmdCheckout.Run(); err != nil {
		return fmt.Errorf("checkout version: %s", err)
	}

	cmdRemoveGit := exec.Command("rm", "-rf", ".git")
	switch runtime.GOOS {
	case "windows":
		cmdRemoveGit = exec.Command("rd", "/s", "/q", ".git")
	}
	cmdRemoveGit.Dir = dir
	if err := cmdRemoveGit.Run(); err != nil {
		return fmt.Errorf("rm -rf .git: %s", err)
	}

	cmdInitGit := exec.Command("git", "init")
	cmdInitGit.Dir = dir
	if err := cmdInitGit.Run(); err != nil {
		return fmt.Errorf("init store: %s", err)
	}

	return nil
}

func GetRepoVersion(language string, version string) (string, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/lhdhtrc/%s/releases/%s", GetTemplateRepo(language), version))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	var data GithubRepoVersion
	body, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(body, &data)
	return data.TagName, nil
}

func GetRepo(cli string, project, language string, version string) error {
	var err error
	if version == "" || version == "latest" {
		version, err = GetRepoVersion(language, "latest")
		if err != nil {
			return err
		}
	}

	var cacheDir string
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		return err
	}

	templateCacheDir := fmt.Sprintf("%s/%s/template/%s", cacheDir, cli, version)
	tempCacheDir := fmt.Sprintf("%s/%s/template/temp/%s", cacheDir, cli, project)

	template, err := os.Stat(templateCacheDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err = GetRepoLocal(language, version, templateCacheDir); err != nil {
			return err
		}
	} else if !template.IsDir() {
		return fmt.Errorf("%s is not a directory", templateCacheDir)
	}

	// 复制模板到临时目录
	if err = CopyDir(templateCacheDir, tempCacheDir); err != nil {
		return err
	}

	return nil
}

//func GetRepo() {
//	var version string
//	GetRepoVersion("", version)
//
//	repo := "https://github.com/lhdhtrc/microservice-go"
//	language := "Go"
//
//	cacheDir := fmt.Sprintf("./.firefly/cache/template/%s/%", strings.ToLower(language), "")
//
//	cmd := exec.Command("git", "clone", "--depth=1", repo, cacheDir)
//	if err := cmd.Run(); err != nil {
//		panic(err)
//	}
//}
