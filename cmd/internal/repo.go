package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetTemplateRepo 获取模版仓库
func GetTemplateRepo(language string) string {
	switch language {
	case "Go":
		return "microservice-go"
	case "NodeJS":
		return "microservice-node"
	case "Rust":
		return "microservice-rust"
	default:
		return ""
	}
}

// GetRepoUrl 获取仓库地址
func GetRepoUrl(language string) string {
	return fmt.Sprintf("https://github.com/lhdhtrc/%s.git", GetTemplateRepo(language))
}

// GetRepoToLocal 获取仓库到本地
func GetRepoToLocal(language string, version string, dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory: %s", err)
	}

	cmd := exec.Command("git", "clone", GetRepoUrl(language), dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %s", err)
	}

	cmdCheckout := exec.Command("git", "checkout", version, "--force")
	cmdCheckout.Dir = dir
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
		return fmt.Errorf("remove .git: %s", err)
	}

	cmdInitGit := exec.Command("git", "init")
	cmdInitGit.Dir = dir
	if err := cmdInitGit.Run(); err != nil {
		return fmt.Errorf("init git: %s", err)
	}

	return nil
}

// GetRepoVersion 获取仓库版本
func GetRepoVersion(language string, version string) (string, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/lhdhtrc/%s/releases/%s", GetTemplateRepo(language), version))
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	var data GithubRepoVersion
	body, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(body, &data)
	return data.TagName, nil
}

// GetRepo 获取仓库
func GetRepo(cli, project, language, version string) error {
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

	templateCacheDir := filepath.Join(cacheDir, cli, "template", version)
	tempCacheDir := filepath.Join(cacheDir, cli, "template", "temp", project)

	templateInfo, err := os.Stat(templateCacheDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err = GetRepoToLocal(language, version, templateCacheDir); err != nil {
			return err
		}
	} else if !templateInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", templateCacheDir)
	}

	if err = CopyDir(templateCacheDir, tempCacheDir); err != nil {
		return err
	}

	return nil
}
