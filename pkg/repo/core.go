package repo

import (
	"encoding/json"
	"fmt"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/file"
	"github.com/fireflycore/cli/pkg/store"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type ConfigEntity struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Project  string `json:"project"`
}

type CoreEntity struct {
	*ConfigEntity

	api  string
	repo string

	currentVersionTemplateCacheDir string
	currentProjectTempDir          string
}

func New(cfg *ConfigEntity) (*CoreEntity, error) {
	core := &CoreEntity{
		ConfigEntity: cfg,
	}
	core.api = fmt.Sprintf("https://api.github.com/repos/%s", config.REPO_OWNER)
	core.repo = fmt.Sprintf("https://github.com/%s/%s.git", config.REPO_OWNER, core.GetTemplate())

	if core.Version == "" || core.Version == "latest" || store.Use.Config.Global.Version[cfg.Language] == "latest" {
		version, err := core.GetVersion()
		if err != nil {
			return nil, err
		}
		store.Use.Config.Global.Version[cfg.Language] = version
		if err = store.Use.Config.UpdateGlobalConfig(); err != nil {
			return nil, err
		}
		core.Version = version
	}

	core.currentProjectTempDir = filepath.Join(store.Use.Config.CacheDir, "temp", core.Project)
	core.currentVersionTemplateCacheDir = filepath.Join(store.Use.Config.CacheTemplateDir, core.Version)

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
	cmd := exec.Command("git", "clone", core.repo, core.currentVersionTemplateCacheDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %s", err)
	}

	cmdCheckout := exec.Command("git", "checkout", core.Version, "--force")
	cmdCheckout.Dir = core.currentVersionTemplateCacheDir
	if err := cmdCheckout.Run(); err != nil {
		return fmt.Errorf("checkout version: %s", err)
	}

	if err := os.RemoveAll(filepath.Join(core.currentVersionTemplateCacheDir, ".git")); err != nil {
		return err
	}

	cmdInitGit := exec.Command("git", "init")
	cmdInitGit.Dir = core.currentVersionTemplateCacheDir
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
	_, err := os.Stat(core.currentVersionTemplateCacheDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err = core.RemoteToLocal(); err != nil {
			return err
		}
	}

	if err = file.CopyDir(core.currentVersionTemplateCacheDir, core.currentProjectTempDir); err != nil {
		return err
	}

	return nil
}
