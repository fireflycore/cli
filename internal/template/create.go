package template

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fireflycore/cli/internal/fsutil"
)

const (
	// GoLanguage 是当前 CLI 支持的模板语言。
	GoLanguage = "go"
	// GoLayoutRepo 是 Firefly Go 服务模板仓库。
	GoLayoutRepo = "https://github.com/fireflycore/go-layout.git"
	// LatestVersion 表示使用模板仓库最新 tag。
	LatestVersion = "latest"
)

// projectNameRegexp 定义 create 允许的项目目录名格式。
var projectNameRegexp = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// CreateOptions 表示 create 命令创建服务时需要的参数。
type CreateOptions struct {
	// Root 是项目生成目标所在目录。
	Root string
	// CacheDir 是模板缓存根目录。
	CacheDir string
	// Name 是项目目录名。
	Name string
	// Language 是模板语言，目前只支持 go。
	Language string
	// TemplateVersion 是模板 tag 或 latest。
	TemplateVersion string
	// Module 是 Go module 名。
	Module string
	// AppID 是写入 bootstrap.json 的 app.id。
	AppID string
	// Service 是写入 bootstrap.json 的 service.name。
	Service string
}

// CreateResult 表示 create 命令完成后的结果。
type CreateResult struct {
	// Name 是创建的项目名。
	Name string
	// Path 是最终项目目录。
	Path string
	// Module 是最终写入的 Go module。
	Module string
	// TemplateVersion 是实际使用的模板版本。
	TemplateVersion string
	// TemplateRepo 是模板仓库地址。
	TemplateRepo string
}

// Create 基于 go-layout 模板创建 Firefly Go 服务项目。
func Create(ctx context.Context, opts CreateOptions) (*CreateResult, error) {
	// 规范化并校验输入。
	opts = normalizeCreateOptions(opts)
	if err := validateCreateOptions(opts); err != nil {
		return nil, err
	}
	// 解析 latest 到真实 tag。
	version, err := ResolveVersion(ctx, opts.TemplateVersion)
	if err != nil {
		return nil, err
	}
	// 确保模板已经缓存到本地。
	templateDir, err := EnsureCached(ctx, opts.CacheDir, version)
	if err != nil {
		return nil, err
	}
	// 复制模板到目标项目目录。
	target := filepath.Join(opts.Root, opts.Name)
	if err = fsutil.CopyDir(templateDir, target); err != nil {
		return nil, err
	}
	// 替换模板项目中的服务信息。
	if err = rewriteGoProject(target, opts, version); err != nil {
		return nil, err
	}
	return &CreateResult{
		Name:            opts.Name,
		Path:            target,
		Module:          opts.Module,
		TemplateVersion: version,
		TemplateRepo:    GoLayoutRepo,
	}, nil
}

// normalizeCreateOptions 补齐 create 参数默认值。
func normalizeCreateOptions(opts CreateOptions) CreateOptions {
	// 目录名和标识字段统一去掉首尾空白。
	opts.Name = strings.TrimSpace(opts.Name)
	opts.Module = strings.TrimSpace(opts.Module)
	opts.AppID = strings.TrimSpace(opts.AppID)
	opts.Service = strings.TrimSpace(opts.Service)
	// 语言默认 go。
	opts.Language = strings.ToLower(strings.TrimSpace(opts.Language))
	if opts.Language == "" {
		opts.Language = GoLanguage
	}
	// 模板版本默认 latest。
	opts.TemplateVersion = strings.TrimSpace(opts.TemplateVersion)
	if opts.TemplateVersion == "" {
		opts.TemplateVersion = LatestVersion
	}
	// module 默认等于项目名。
	if opts.Module == "" {
		opts.Module = opts.Name
	}
	// app id 默认等于项目名。
	if opts.AppID == "" {
		opts.AppID = opts.Name
	}
	// service 默认等于项目名。
	if opts.Service == "" {
		opts.Service = opts.Name
	}
	return opts
}

// validateCreateOptions 校验 create 参数是否可以创建项目。
func validateCreateOptions(opts CreateOptions) error {
	// 目标根目录必须存在。
	if opts.Root == "" {
		return fmt.Errorf("root is required")
	}
	// 缓存目录必须可推导。
	if opts.CacheDir == "" {
		return fmt.Errorf("cache dir is required")
	}
	// 项目名不能为空且必须适合作为目录名。
	if opts.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if !projectNameRegexp.MatchString(opts.Name) {
		return fmt.Errorf("invalid project name %q", opts.Name)
	}
	// 当前只支持 Go 模板。
	if opts.Language != GoLanguage {
		return fmt.Errorf("unsupported language %q", opts.Language)
	}
	return nil
}

// ResolveVersion 将 latest 解析为 go-layout 最新 tag。
func ResolveVersion(ctx context.Context, version string) (string, error) {
	// 显式 tag 直接返回。
	version = strings.TrimSpace(version)
	if version != "" && version != LatestVersion {
		return version, nil
	}
	// 读取远端 tag 列表。
	output, err := runGit(ctx, "", "ls-remote", "--tags", "--refs", GoLayoutRepo)
	if err != nil {
		return "", err
	}
	// 提取 tag 名。
	tags := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.HasPrefix(fields[1], "refs/tags/") {
			tags = append(tags, strings.TrimPrefix(fields[1], "refs/tags/"))
		}
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in %s", GoLayoutRepo)
	}
	// 按语义化版本排序，最后一个就是最新。
	sort.Slice(tags, func(i, j int) bool {
		return compareVersion(tags[i], tags[j]) < 0
	})
	return tags[len(tags)-1], nil
}

// EnsureCached 确保指定版本模板存在于本地缓存。
func EnsureCached(ctx context.Context, cacheDir, version string) (string, error) {
	// 模板缓存按仓库名和版本隔离。
	target := filepath.Join(cacheDir, "templates", "go-layout", version)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return target, nil
	}
	// 创建父目录。
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", err
	}
	// 使用临时目录避免半截 clone 污染正式缓存。
	tmp := target + ".tmp-" + strconv.Itoa(os.Getpid())
	_ = os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)
	// 克隆指定 tag。
	if _, err := runGit(ctx, "", "clone", "--depth", "1", "--branch", version, GoLayoutRepo, tmp); err != nil {
		return "", err
	}
	// 删除模板自身 git 元数据。
	if err := os.RemoveAll(filepath.Join(tmp, ".git")); err != nil {
		return "", err
	}
	// 原子迁移到正式缓存路径。
	if err := os.Rename(tmp, target); err != nil {
		return "", err
	}
	return target, nil
}

// rewriteGoProject 将模板内容改写为目标服务信息。
func rewriteGoProject(root string, opts CreateOptions, version string) error {
	// 读取模板原始 module，用于替换 import path。
	oldModule, err := fsutil.ReadModule(root)
	if err != nil {
		return err
	}
	// 写入目标 module。
	if err = fsutil.WriteModule(root, opts.Module); err != nil {
		return err
	}
	// 执行文本替换，不碰 git 元数据和锁文件。
	replacements := map[string]string{
		oldModule:   opts.Module,
		"go-layout": opts.Name,
	}
	skip := map[string]bool{
		".git":       true,
		".github":    true,
		"go.sum":     true,
		"LICENSE":    true,
		"project.md": true,
	}
	if err = fsutil.ReplaceInTextFiles(root, replacements, skip); err != nil {
		return err
	}
	// 写入 bootstrap.json 中的服务身份。
	bootstrapPath := filepath.Join(root, "conf", "bootstrap.json")
	if _, err = os.Stat(bootstrapPath); err == nil {
		err = fsutil.UpdateJSONFile(bootstrapPath, func(raw map[string]any) {
			fsutil.SetNestedString(raw, []string{"app", "id"}, opts.AppID)
			fsutil.SetNestedString(raw, []string{"service", "name"}, opts.Service)
		})
		if err != nil {
			return err
		}
	}
	// 用项目专属 README 覆盖模板 README。
	return writeReadme(root, opts, version)
}

// writeReadme 写入生成项目的 README。
func writeReadme(root string, opts CreateOptions, version string) error {
	// README 属于生成项目文档，使用中文方便团队维护。
	content := fmt.Sprintf(`# %s 服务

## 项目说明


## 基础信息

- language: %s
- template: %s
- module: %s

## 初始化运行

%s
go mod tidy
buf generate
go run main.go
%s
`, opts.Name, opts.Language, version, opts.Module, "```bash", "```")
	return os.WriteFile(filepath.Join(root, "README.md"), []byte(content), 0644)
}

// runGit 执行 git 命令并返回标准输出。
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	// 构造 git 子进程。
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	// 捕获 stderr 便于错误提示。
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return string(output), nil
}

// compareVersion 比较 v0.3.1 这类版本号。
func compareVersion(left, right string) int {
	// 解析为整数切片。
	lv := parseVersion(left)
	rv := parseVersion(right)
	// 逐段比较主版本、次版本和修订版本。
	for index := 0; index < 3; index++ {
		if lv[index] < rv[index] {
			return -1
		}
		if lv[index] > rv[index] {
			return 1
		}
	}
	// 数字段相同时用原始字符串稳定排序。
	return strings.Compare(left, right)
}

// parseVersion 将版本号解析成三段数字。
func parseVersion(version string) [3]int {
	// 去掉常见 v 前缀。
	version = strings.TrimPrefix(version, "v")
	// 只解析前三段。
	parts := strings.Split(version, ".")
	var result [3]int
	for index := 0; index < len(parts) && index < 3; index++ {
		value, _ := strconv.Atoi(parts[index])
		result[index] = value
	}
	return result
}
