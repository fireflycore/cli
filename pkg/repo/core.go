package repo

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/file"
	"github.com/fireflycore/cli/pkg/store"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// README 是生成业务项目时写入的 README 模板内容。
//
//go:embed README.md
var README []byte

// ConfigEntity 表示 create 命令传给模板仓库模块的创建参数。
type ConfigEntity struct {
	// Language 是目标开发语言，目前主线只支持 go。
	Language string `json:"language"`
	// Version 是模板仓库 tag，latest 会被解析成最新 release。
	Version string `json:"version"`
	// Project 是要生成的项目目录名。
	Project string `json:"project"`
	// Module 是 Go 项目的 module 名，未传入时使用 Project。
	Module string `json:"module"`
	// AppID 是写入 bootstrap.json 的 Firefly app.id。
	AppID string `json:"app_id"`
	// Service 是写入 bootstrap.json 的 service.name。
	Service string `json:"service"`
}

// CoreEntity 封装模板仓库拉取、缓存和项目初始化流程。
type CoreEntity struct {
	// ConfigEntity 保存创建项目时传入的业务参数。
	*ConfigEntity

	// api 是 GitHub release 查询接口前缀。
	api string
	// repo 是模板仓库 git clone 地址。
	repo string

	// currentVersionTemplateCacheDir 是当前模板版本的本地缓存目录。
	currentVersionTemplateCacheDir string
	// currentProjectTempDir 是生成项目时使用的临时目录。
	currentProjectTempDir string
}

// New 根据创建参数初始化模板仓库 CoreEntity。
func New(cfg *ConfigEntity) (*CoreEntity, error) {
	// 先保存外部传入的创建参数。
	core := &CoreEntity{
		ConfigEntity: cfg,
	}
	// GitHub API 只用于查询模板最新 release。
	core.api = fmt.Sprintf("https://api.github.com/repos/%s", config.REPO_OWNER)
	// 模板仓库地址由语言映射出的模板名决定。
	core.repo = fmt.Sprintf("https://github.com/%s/%s.git", config.REPO_OWNER, core.GetTemplate())

	// latest 或空版本都需要解析成真实 release tag，避免缓存目录不稳定。
	if core.Version == "" || core.Version == "latest" || store.Use.Config.Global.Version[cfg.Language] == "latest" {
		// 从 GitHub release 读取最新模板版本。
		version, err := core.GetVersion()
		if err != nil {
			return nil, err
		}
		// 将解析到的版本写回全局配置，下一次创建可以复用。
		store.Use.Config.Global.Version[cfg.Language] = version
		if err = store.Use.Config.UpdateGlobalConfig(); err != nil {
			return nil, err
		}
		// 当前创建流程也使用真实版本号。
		core.Version = version
	}

	// 临时目录用于生成项目，最后再复制到用户当前目录。
	core.currentProjectTempDir = filepath.Join(store.Use.Config.CacheDir, "temp", core.Project)
	// 模板缓存目录按版本隔离，避免不同模板版本互相覆盖。
	core.currentVersionTemplateCacheDir = filepath.Join(store.Use.Config.CacheTemplateDir, core.Version)

	return core, nil
}

// GetTemplate 根据语言返回模板仓库名。
func (core *CoreEntity) GetTemplate() string {
	// 当前只有 Go 模板接入 Firefly CLI 主线。
	switch core.Language {
	case "go":
		return "go-layout"
	default:
		return ""
	}
}

// RemoteToLocal 将远端模板仓库拉取到本地版本缓存目录。
func (core *CoreEntity) RemoteToLocal() error {
	// 克隆模板仓库到当前版本缓存目录。
	cmd := exec.Command("git", "clone", core.repo, core.currentVersionTemplateCacheDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %s", err)
	}

	// 切换到指定模板版本，确保生成内容稳定可复现。
	cmdCheckout := exec.Command("git", "checkout", core.Version, "--force")
	// checkout 必须在模板缓存目录内执行。
	cmdCheckout.Dir = core.currentVersionTemplateCacheDir
	if err := cmdCheckout.Run(); err != nil {
		return fmt.Errorf("checkout version: %s", err)
	}

	// 删除模板仓库原始 git 信息，避免生成项目继承模板仓库历史。
	if err := os.RemoveAll(filepath.Join(core.currentVersionTemplateCacheDir, ".git")); err != nil {
		return err
	}

	// 重新初始化 git，让缓存目录后续保持普通项目形态。
	cmdInitGit := exec.Command("git", "init")
	// git init 同样在模板缓存目录内执行。
	cmdInitGit.Dir = core.currentVersionTemplateCacheDir
	if err := cmdInitGit.Run(); err != nil {
		return fmt.Errorf("init git: %s", err)
	}

	return nil
}

// GetVersion 从 GitHub release API 读取模板版本。
func (core *CoreEntity) GetVersion() (string, error) {
	// release 路径中 latest 会被 GitHub 解析成最新发布版本。
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/releases/%s", core.api, core.GetTemplate(), core.Version), nil)
	if err != nil {
		return "", err
	}
	// 如果配置了仓库 token，则带上授权头以降低 GitHub API 限流影响。
	if len(config.REPO_TOKEN) != 0 {
		req.Header.Set("Authorization", config.REPO_TOKEN)
	}

	// 使用默认 HTTP client 访问 GitHub API。
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// 读取完成后关闭响应体。
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// 非 200 响应统一视为 release 查询失败。
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api limit restrict")
	}

	// GitHub release 响应只需要 tag_name。
	var data GithubRepoVersion
	// 读取响应体并反序列化 release 信息。
	body, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(body, &data)

	// 返回模板版本 tag。
	return data.TagName, nil
}

// GetRepo 准备模板并初始化最终项目目录。
func (core *CoreEntity) GetRepo() error {
	// 检查当前版本模板是否已经存在本地缓存。
	_, err := os.Stat(core.currentVersionTemplateCacheDir)
	if err != nil {
		// 只有“不存在”需要拉取远端模板，其他 stat 错误直接返回。
		if !os.IsNotExist(err) {
			return err
		}

		// 本地没有缓存时从远端克隆模板。
		if err = core.RemoteToLocal(); err != nil {
			return err
		}
	}

	// 将版本缓存复制到当前项目临时目录，后续替换都在临时目录完成。
	if err = file.CopyDir(core.currentVersionTemplateCacheDir, core.currentProjectTempDir); err != nil {
		return err
	}

	// 基于临时目录执行项目初始化。
	return core.InitProject()
}

// InitProject 根据语言执行模板替换并写入最终项目目录。
func (core *CoreEntity) InitProject() error {
	// 不同语言的模板替换策略不同，当前只处理 Go。
	switch core.Language {
	case "go":
		// 将模板中的项目名替换为用户指定的项目名。
		if err := file.WalkDirAndReplace(core.Language, core.currentProjectTempDir, core.GetTemplate(), core.Project); err != nil {
			return err
		}
		// 更新 go.mod 和 import path 中的 module 名。
		if err := core.applyGoModule(); err != nil {
			return err
		}
		// 按参数写入 bootstrap.json 中的 app.id 和 service.name。
		if err := core.updateBootstrap(); err != nil {
			return err
		}
		// 兼容模板 run.sh 中的 project_name 占位符。
		_ = file.ReplaceInFile(filepath.Join(core.currentProjectTempDir, "run.sh"), `"project_name"`, fmt.Sprintf(`"%s"`, core.Project))
		// 写入基于当前项目参数渲染后的 README。
		if err := core.WriteReadme(); err != nil {
			return err
		}
		// 将临时项目复制到用户当前目录下的项目目录。
		if err := file.CopyDir(core.currentProjectTempDir, filepath.Join(store.Use.Config.LocalDir, core.Project)); err != nil {
			return err
		}
		// 清理临时目录，避免下次 create 拿到旧内容。
		return os.RemoveAll(core.currentProjectTempDir)
	}
	// 暂未支持的语言不做处理。
	return nil
}

// WriteReadme 渲染并写入新项目 README.md。
func (core *CoreEntity) WriteReadme() error {
	// 解析内嵌 README 模板。
	tmpl, err := template.New("README").Parse(string(README))
	if err != nil {
		return err
	}

	// 准备模板渲染所需的项目数据。
	data := ReadmeEntity{
		Project:  core.Project,
		Language: core.Language,
		Version:  core.Version,
		Module:   core.moduleName(),
	}

	// 使用内存缓冲区暂存渲染结果。
	var buf bytes.Buffer

	// 执行模板渲染。
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	// 在临时项目目录中创建 README.md。
	outputFile, err := os.Create(filepath.Join(core.currentProjectTempDir, "README.md"))
	if err != nil {
		return err
	}
	// 写入完成后关闭文件。
	defer outputFile.Close()

	// 将渲染结果写入 README.md。
	_, err = buf.WriteTo(outputFile)
	if err != nil {
		return err
	}
	return nil
}

// applyGoModule 将 Go 模板中的 module 和 import path 替换成目标 module。
func (core *CoreEntity) applyGoModule() error {
	// 优先使用显式 module，未传入时使用项目名。
	module := core.moduleName()
	// module 与项目名相同则无需替换。
	if module == "" || module == core.Project {
		return nil
	}

	// 替换 go.mod 的 module 声明。
	if err := file.ReplaceInFile(
		filepath.Join(core.currentProjectTempDir, "go.mod"),
		fmt.Sprintf("module %s", core.Project),
		fmt.Sprintf("module %s", module),
	); err != nil {
		return err
	}

	// 替换 Go 代码或配置中带引号的 import path 前缀。
	replacements := map[string]string{
		fmt.Sprintf("\"%s/", core.Project): fmt.Sprintf("\"%s/", module),
		fmt.Sprintf("'%s/", core.Project):  fmt.Sprintf("'%s/", module),
	}
	// 对每组替换规则执行全目录扫描，忽略目录由 file 包统一处理。
	for oldText, newText := range replacements {
		if err := file.WalkDirAndReplace(core.Language, core.currentProjectTempDir, oldText, newText); err != nil {
			return err
		}
	}
	return nil
}

// updateBootstrap 按 create 参数更新模板项目的 conf/bootstrap.json。
func (core *CoreEntity) updateBootstrap() error {
	// bootstrap.json 是 go-layout 模板中的启动配置文件。
	path := filepath.Join(core.currentProjectTempDir, "conf", "bootstrap.json")
	// 模板中没有 bootstrap.json 时保持兼容，直接跳过。
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// 使用 map 保留未知字段，只改 CLI 关心的路径。
	var raw map[string]any
	if err = json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// app id 非空时写入 app.id。
	if core.AppID != "" {
		setNested(raw, []string{"app", "id"}, core.AppID)
	}
	// service 非空时写入 service.name。
	if core.Service != "" {
		setNested(raw, []string{"service", "name"}, core.Service)
	}

	// 重新格式化 JSON，保持配置文件可读。
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	// 写回时补一个换行，符合常见文本文件习惯。
	out = append(out, '\n')
	// 将更新后的 bootstrap.json 写回临时项目。
	return os.WriteFile(path, out, 0644)
}

// moduleName 返回最终写入项目的 Go module 名。
func (core *CoreEntity) moduleName() string {
	// 显式传入 module 时优先使用。
	if core.Module != "" {
		return core.Module
	}
	// 未传入 module 时沿用项目名。
	return core.Project
}

// setNested 在 map 中按路径写入嵌套字段。
func setNested(raw map[string]any, path []string, value string) {
	// 空路径不需要写入。
	if len(path) == 0 {
		return
	}
	// 路径只剩最后一段时直接赋值。
	if len(path) == 1 {
		raw[path[0]] = value
		return
	}

	// 读取下一层对象，类型不匹配时创建新 map 覆盖。
	next, ok := raw[path[0]].(map[string]any)
	if !ok {
		next = make(map[string]any)
		raw[path[0]] = next
	}
	// 递归写入剩余路径。
	setNested(next, path[1:], value)
}
