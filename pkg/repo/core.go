package repo

import (
	"encoding/json"
	"fmt"
	"github.com/fireflycore/cli/pkg/file"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type ConfigEntity struct {
	Dir      string `json:"dir"`
	Owner    string `json:"owner"`
	Language string `json:"language"`
	Version  string `json:"version"`
	Project  string `json:"project"`
}

type CoreEntity struct {
	*ConfigEntity

	api                     string
	repo                    string
	cacheDir                string
	configDir               string
	templateCacheDir        string
	currentTemplateCacheDir string

	tempProjectDir string
}

func New(config *ConfigEntity) (*CoreEntity, error) {
	core := &CoreEntity{
		ConfigEntity: config,
	}
	core.api = fmt.Sprintf("https://api.github.com/repos/%s", core.Owner)
	core.repo = fmt.Sprintf("https://github.com/%s/%s.git", config.Owner, core.GetTemplate())

	if core.Version == "" || core.Version == "latest" {
		version, err := core.GetVersion()
		if err != nil {
			return nil, err
		}
		core.Version = version
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	core.cacheDir = filepath.Join(cache, core.Dir, "cache")
	core.configDir = filepath.Join(cache, core.Dir, "config")
	core.templateCacheDir = filepath.Join(core.cacheDir, "template")
	core.currentTemplateCacheDir = filepath.Join(core.templateCacheDir, core.Version)

	core.tempProjectDir = filepath.Join(core.cacheDir, "temp", core.Project)

	fmt.Println(core.cacheDir)
	fmt.Println(core.currentTemplateCacheDir)
	fmt.Println(core.tempProjectDir)

	return core, nil
}

// GetTemplate 获取模版
func (core *CoreEntity) GetTemplate() string {
	switch core.Language {
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

// RemoteToLocal 获取到本地
func (core *CoreEntity) RemoteToLocal() error {
	cmd := exec.Command("git", "clone", core.repo, core.currentTemplateCacheDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %s", err)
	}

	cmdCheckout := exec.Command("git", "checkout", core.Version, "--force")
	cmdCheckout.Dir = core.currentTemplateCacheDir
	if err := cmdCheckout.Run(); err != nil {
		return fmt.Errorf("checkout version: %s", err)
	}

	cmdRemoveGit := exec.Command("rm", "-rf", ".git")
	switch runtime.GOOS {
	case "windows":
		cmdRemoveGit = exec.Command("rd", "/s", "/q", ".git")
	}
	cmdRemoveGit.Dir = core.currentTemplateCacheDir
	if err := cmdRemoveGit.Run(); err != nil {
		return fmt.Errorf("remove .git: %s", err)
	}

	cmdInitGit := exec.Command("git", "init")
	cmdInitGit.Dir = core.currentTemplateCacheDir
	if err := cmdInitGit.Run(); err != nil {
		return fmt.Errorf("init git: %s", err)
	}

	return nil
}

// GetVersion 获取仓库版本
func (core *CoreEntity) GetVersion() (string, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s/releases/%s", core.api, core.GetTemplate(), core.Version))
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api limit restrict")
	}

	var data GithubRepoVersion
	body, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(body, &data)

	return data.TagName, nil
}

// GetRepo 获取仓库
func (core *CoreEntity) GetRepo() error {
	_, err := os.Stat(core.currentTemplateCacheDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err = core.RemoteToLocal(); err != nil {
			return err
		}
	}

	if err = file.CopyDir(core.currentTemplateCacheDir, core.tempProjectDir); err != nil {
		return err
	}

	return nil
}
