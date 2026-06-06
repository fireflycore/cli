package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fireflycore/cli/internal/fsutil"
	"gopkg.in/yaml.v3"
)

const (
	// ConfigDirName 是项目元信息目录名。
	ConfigDirName = ".firefly"
	// ConfigFileName 是项目元信息文件名。
	ConfigFileName = "project.yaml"
	// SchemaVersion 是当前 project.yaml schema 版本。
	SchemaVersion = "firefly.project.v1"

	// DefaultLanguage 是默认项目语言。
	DefaultLanguage = "go"
	// DefaultNamespace 是默认服务命名空间。
	DefaultNamespace = "lhdht"
	// DefaultBootstrapFile 是默认启动配置路径。
	DefaultBootstrapFile = "conf/bootstrap.json"
	// DefaultBootstrapVersionPath 是默认版本字段路径，对应 bootstrapConf.app.version。
	DefaultBootstrapVersionPath = "app.version"
	// DefaultDescriptorDir 是 descriptor 默认输出目录。
	DefaultDescriptorDir = "dist/descriptors"
	// DefaultDescriptorFileTemplate 是本地 descriptor 文件名模板。
	DefaultDescriptorFileTemplate = "${version}.pb"
	// DefaultObjectKeyTemplate 是 S3 对象 key 模板。
	DefaultObjectKeyTemplate = "${service}/${version}.pb"
	// DefaultDescriptorContentType 是 descriptor 上传时的默认 content type。
	DefaultDescriptorContentType = "application/octet-stream"
	// DefaultS3Region 是 S3 SDK 需要的默认 region。
	DefaultS3Region = "us-east-1"
	// DefaultS3Bucket 是 descriptor 主线使用的默认 bucket。
	DefaultS3Bucket = "descriptor"
)

// Config 对应 .firefly/project.yaml 的完整结构。
type Config struct {
	// Schema 标识配置文件格式版本。
	Schema string `yaml:"schema" json:"schema"`
	// Service 保存服务身份信息。
	Service ServiceConfig `yaml:"service" json:"service"`
	// Bootstrap 保存服务版本读取规则。
	Bootstrap BootstrapConfig `yaml:"bootstrap" json:"bootstrap"`
	// Descriptor 保存 descriptor 本地和远端路径规则。
	Descriptor DescriptorConfig `yaml:"descriptor" json:"descriptor"`
	// S3 保存 S3 兼容对象存储配置。
	S3 S3Config `yaml:"s3" json:"s3"`
}

// ServiceConfig 保存服务维度的基础元信息。
type ServiceConfig struct {
	// Name 是服务名。
	Name string `yaml:"name" json:"name"`
	// AppID 是 Firefly app id。
	AppID string `yaml:"app_id" json:"app_id"`
	// Namespace 是服务命名空间。
	Namespace string `yaml:"namespace" json:"namespace"`
	// Language 是项目语言。
	Language string `yaml:"language" json:"language"`
	// Module 是 Go module 名。
	Module string `yaml:"module" json:"module"`
}

// BootstrapConfig 描述如何从启动配置读取服务版本。
type BootstrapConfig struct {
	// File 是启动配置文件路径。
	File string `yaml:"file" json:"file"`
	// VersionPath 是版本字段路径。
	VersionPath string `yaml:"version_path" json:"version_path"`
}

// DescriptorConfig 描述 descriptor 的文件和 URL 推导规则。
type DescriptorConfig struct {
	// Dir 是本地 descriptor 输出目录。
	Dir string `yaml:"dir" json:"dir"`
	// FileTemplate 是本地 descriptor 文件名模板。
	FileTemplate string `yaml:"file_template" json:"file_template"`
	// ObjectKeyTemplate 是 S3 object key 模板。
	ObjectKeyTemplate string `yaml:"object_key_template" json:"object_key_template"`
	// RefTemplate 是 descriptor_ref URL 模板。
	RefTemplate string `yaml:"descriptor_ref_template,omitempty" json:"descriptor_ref_template,omitempty"`
	// Ref 是固定 descriptor_ref。
	Ref string `yaml:"descriptor_ref,omitempty" json:"descriptor_ref,omitempty"`
	// ContentType 是上传 descriptor 时使用的 content type。
	ContentType string `yaml:"content_type" json:"content_type"`
}

// S3Config 保存 S3 兼容对象存储配置。
type S3Config struct {
	// Profile 是 AWS shared config profile。
	Profile string `yaml:"profile" json:"profile"`
	// Region 是 S3 region。
	Region string `yaml:"region" json:"region"`
	// Endpoint 是 S3 兼容 endpoint。
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// Bucket 是目标 bucket。
	Bucket string `yaml:"bucket" json:"bucket"`
	// ForcePathStyle 控制是否使用 path-style 地址。
	ForcePathStyle bool `yaml:"force_path_style" json:"force_path_style"`
}

// InitOptions 是 project init 命令的输入参数。
type InitOptions struct {
	// Root 是项目根目录。
	Root string
	// ServiceName 是服务名。
	ServiceName string
	// AppID 是 Firefly app id。
	AppID string
	// Namespace 是服务命名空间。
	Namespace string
	// Language 是项目语言。
	Language string
	// Module 是 Go module 名。
	Module string
	// BootstrapFile 是启动配置文件路径。
	BootstrapFile string
	// VersionPath 是版本字段路径。
	VersionPath string
	// DescriptorDir 是 descriptor 输出目录。
	DescriptorDir string
	// FileTemplate 是 descriptor 文件名模板。
	FileTemplate string
	// ObjectKeyTemplate 是 S3 object key 模板。
	ObjectKeyTemplate string
	// DescriptorRef 是固定 descriptor_ref。
	DescriptorRef string
	// RefTemplate 是 descriptor_ref URL 模板。
	RefTemplate string
	// ContentType 是上传 content type。
	ContentType string
	// S3Profile 是 AWS shared config profile。
	S3Profile string
	// S3Region 是 S3 region。
	S3Region string
	// S3Endpoint 是 S3 兼容 endpoint。
	S3Endpoint string
	// S3Bucket 是目标 bucket。
	S3Bucket string
	// ForcePathStyle 控制是否使用 path-style 地址。
	ForcePathStyle bool
	// Overwrite 控制是否覆盖已有 project.yaml。
	Overwrite bool
}

// Resolved 保存基于配置和版本推导出的发布信息。
type Resolved struct {
	// ConfigPath 是 project.yaml 路径。
	ConfigPath string
	// Version 是从 bootstrapConf.app.version 读取到的服务版本。
	Version string
	// DescriptorFile 是本地 descriptor 文件路径。
	DescriptorFile string
	// ObjectKey 是 S3 object key。
	ObjectKey string
	// DescriptorRef 是 gateway 可读取的 descriptor URL。
	DescriptorRef string
}

// CheckStatus 是本地检查的单项状态。
type CheckStatus string

const (
	// CheckOK 表示检查通过。
	CheckOK CheckStatus = "ok"
	// CheckWarn 表示检查存在提醒但不阻断。
	CheckWarn CheckStatus = "warn"
	// CheckFailed 表示检查失败并应导致命令返回非零状态。
	CheckFailed CheckStatus = "failed"
)

// CheckResult 表示一条本地检查结果。
type CheckResult struct {
	// Name 是检查项名称。
	Name string
	// Status 是检查项状态。
	Status CheckStatus
	// Message 是检查结果说明。
	Message string
}

// ConfigPath 返回项目配置文件路径。
func ConfigPath(root string) string {
	return filepath.Join(root, ConfigDirName, ConfigFileName)
}

// Load 读取并解析项目配置。
func Load(root string) (*Config, string, error) {
	// 读取 project.yaml。
	path := ConfigPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}
	// 解析 YAML。
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, path, err
	}
	// 补齐默认值。
	cfg.ApplyDefaults()
	// 校验 schema，避免误读其他文件。
	if cfg.Schema != SchemaVersion {
		return nil, path, fmt.Errorf("unsupported project schema %q", cfg.Schema)
	}
	return &cfg, path, nil
}

// Save 写入项目配置。
func Save(root string, cfg *Config, overwrite bool) (string, error) {
	// 计算输出路径。
	path := ConfigPath(root)
	// 默认不覆盖已有配置。
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return path, fmt.Errorf("%s already exists", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return path, err
		}
	}
	// 确保 .firefly 目录存在。
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return path, err
	}
	// 补齐默认值后序列化。
	cfg.ApplyDefaults()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return path, err
	}
	return path, os.WriteFile(path, data, 0644)
}

// NewConfig 根据命令输入构造项目配置。
func NewConfig(opts InitOptions) (*Config, error) {
	// 根目录为空时使用当前工作目录。
	root := opts.Root
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	// 服务名默认使用当前目录名。
	service := firstNonEmpty(opts.ServiceName, filepath.Base(root))
	// module 默认从 go.mod 读取。
	module := strings.TrimSpace(opts.Module)
	if module == "" {
		if value, err := fsutil.ReadModule(root); err == nil {
			module = value
		}
	}
	// app id 默认等于服务名。
	appID := firstNonEmpty(opts.AppID, service)
	// endpoint 存在时默认生成路径风格 descriptor_ref 模板。
	refTemplate := strings.TrimSpace(opts.RefTemplate)
	if refTemplate == "" && strings.TrimSpace(opts.DescriptorRef) == "" && strings.TrimSpace(opts.S3Endpoint) != "" {
		refTemplate = "${endpoint}/${bucket}/${service}/${version}.pb"
	}
	// 组装配置结构。
	cfg := &Config{
		Schema: SchemaVersion,
		Service: ServiceConfig{
			Name:      service,
			AppID:     appID,
			Namespace: firstNonEmpty(opts.Namespace, DefaultNamespace),
			Language:  strings.ToLower(firstNonEmpty(opts.Language, DefaultLanguage)),
			Module:    module,
		},
		Bootstrap: BootstrapConfig{
			File:        firstNonEmpty(opts.BootstrapFile, DefaultBootstrapFile),
			VersionPath: firstNonEmpty(opts.VersionPath, DefaultBootstrapVersionPath),
		},
		Descriptor: DescriptorConfig{
			Dir:               firstNonEmpty(opts.DescriptorDir, DefaultDescriptorDir),
			FileTemplate:      firstNonEmpty(opts.FileTemplate, DefaultDescriptorFileTemplate),
			ObjectKeyTemplate: firstNonEmpty(opts.ObjectKeyTemplate, DefaultObjectKeyTemplate),
			RefTemplate:       refTemplate,
			Ref:               strings.TrimSpace(opts.DescriptorRef),
			ContentType:       firstNonEmpty(opts.ContentType, DefaultDescriptorContentType),
		},
		S3: S3Config{
			Profile:        strings.TrimSpace(opts.S3Profile),
			Region:         firstNonEmpty(opts.S3Region, DefaultS3Region),
			Endpoint:       strings.TrimSpace(opts.S3Endpoint),
			Bucket:         firstNonEmpty(opts.S3Bucket, DefaultS3Bucket),
			ForcePathStyle: opts.ForcePathStyle,
		},
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

// ApplyDefaults 为可省略字段补默认值。
func (cfg *Config) ApplyDefaults() {
	// schema 为空时写入当前版本。
	if cfg.Schema == "" {
		cfg.Schema = SchemaVersion
	}
	// 服务语言默认 Go。
	if cfg.Service.Language == "" {
		cfg.Service.Language = DefaultLanguage
	}
	// 命名空间默认 lhdht。
	if cfg.Service.Namespace == "" {
		cfg.Service.Namespace = DefaultNamespace
	}
	// app id 默认等于服务名。
	if cfg.Service.AppID == "" {
		cfg.Service.AppID = cfg.Service.Name
	}
	// bootstrap 默认路径。
	if cfg.Bootstrap.File == "" {
		cfg.Bootstrap.File = DefaultBootstrapFile
	}
	// 服务版本默认从 app.version 读取。
	if cfg.Bootstrap.VersionPath == "" {
		cfg.Bootstrap.VersionPath = DefaultBootstrapVersionPath
	}
	// descriptor 默认目录。
	if cfg.Descriptor.Dir == "" {
		cfg.Descriptor.Dir = DefaultDescriptorDir
	}
	// descriptor 默认文件名使用版本号。
	if cfg.Descriptor.FileTemplate == "" {
		cfg.Descriptor.FileTemplate = DefaultDescriptorFileTemplate
	}
	// object key 默认使用 service/version。
	if cfg.Descriptor.ObjectKeyTemplate == "" {
		cfg.Descriptor.ObjectKeyTemplate = DefaultObjectKeyTemplate
	}
	// 上传类型默认二进制。
	if cfg.Descriptor.ContentType == "" {
		cfg.Descriptor.ContentType = DefaultDescriptorContentType
	}
	// region 给 SDK 一个默认值。
	if cfg.S3.Region == "" {
		cfg.S3.Region = DefaultS3Region
	}
	// bucket 默认 descriptor。
	if cfg.S3.Bucket == "" {
		cfg.S3.Bucket = DefaultS3Bucket
	}
}

// Resolve 推导本地 descriptor 文件和远端对象路径。
func (cfg *Config) Resolve(root, configPath string) (*Resolved, error) {
	// 从 bootstrap 中读取服务版本。
	version, err := cfg.ReadVersion(root)
	if err != nil {
		return nil, err
	}
	// 展开模板变量。
	vars := cfg.templateVars(version)
	fileName := expand(cfg.Descriptor.FileTemplate, vars)
	objectKey := filepath.ToSlash(expand(cfg.Descriptor.ObjectKeyTemplate, vars))
	descriptorRef := cfg.ResolveDescriptorRef(version, objectKey)
	return &Resolved{
		ConfigPath:     configPath,
		Version:        version,
		DescriptorFile: filepath.Join(root, cfg.Descriptor.Dir, fileName),
		ObjectKey:      objectKey,
		DescriptorRef:  descriptorRef,
	}, nil
}

// ReadVersion 从 bootstrap 配置读取服务版本。
func (cfg *Config) ReadVersion(root string) (string, error) {
	// 读取启动配置文件。
	path := filepath.Join(root, cfg.Bootstrap.File)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read bootstrap config: %w", err)
	}
	// JSON 和 YAML 都解析成通用结构。
	var raw any
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &raw)
	default:
		err = json.Unmarshal(data, &raw)
	}
	if err != nil {
		return "", fmt.Errorf("parse bootstrap config: %w", err)
	}
	// 按点分路径读取版本。
	value, ok := lookup(raw, strings.Split(cfg.Bootstrap.VersionPath, "."))
	if !ok {
		return "", fmt.Errorf("bootstrap version path %q not found", cfg.Bootstrap.VersionPath)
	}
	version := strings.TrimSpace(fmt.Sprint(value))
	if version == "" || version == "<nil>" {
		return "", fmt.Errorf("bootstrap version path %q is empty", cfg.Bootstrap.VersionPath)
	}
	return version, nil
}

// ResolveDescriptorRef 推导 descriptor_ref。
func (cfg *Config) ResolveDescriptorRef(version, objectKey string) string {
	// 构造模板变量并补充完整 key。
	vars := cfg.templateVars(version)
	vars["key"] = objectKey
	// 模板优先，固定值其次。
	if cfg.Descriptor.RefTemplate != "" {
		return expand(cfg.Descriptor.RefTemplate, vars)
	}
	if cfg.Descriptor.Ref != "" {
		return expand(cfg.Descriptor.Ref, vars)
	}
	// 缺少 endpoint 时无法自动推导。
	if cfg.S3.Endpoint == "" || cfg.S3.Bucket == "" || objectKey == "" {
		return ""
	}
	return strings.TrimRight(cfg.S3.Endpoint, "/") + "/" + cfg.S3.Bucket + "/" + objectKey
}

// ReadFireflyModules 从 go.mod 中提取 Firefly 依赖版本。
func ReadFireflyModules(root string) map[string]string {
	// 读取 go.mod。
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil
	}
	// 逐行识别 github.com/fireflycore/* 依赖。
	deps := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		if strings.HasPrefix(fields[0], "github.com/fireflycore/") {
			deps[fields[0]] = fields[1]
		}
	}
	return deps
}

// Check 执行本地项目检查，不连接运行时组件。
func Check(root string) ([]CheckResult, *Config, *Resolved) {
	// add 统一追加检查结果。
	results := make([]CheckResult, 0, 10)
	add := func(name string, status CheckStatus, message string) {
		results = append(results, CheckResult{Name: name, Status: status, Message: message})
	}
	// 读取项目配置。
	cfg, path, err := Load(root)
	if err != nil {
		add("project config", CheckFailed, err.Error())
		return results, nil, nil
	}
	add("project config", CheckOK, path)
	// 检查 Go module。
	if module, err := fsutil.ReadModule(root); err != nil {
		add("go.mod", CheckWarn, "go.mod not found or module line missing")
	} else {
		add("go.mod", CheckOK, module)
	}
	// 检查基础构建文件，go-layout 模板当前使用小写 makefile。
	checkAnyFile(root, "makefile", []string{"Makefile", "makefile"}, "makefile exists", "Makefile or makefile missing", add)
	checkFile(root, "buf.yaml", "buf.yaml exists", "buf.yaml missing", add)
	checkFile(root, "buf.gen.yaml", "buf.gen.yaml exists", "buf.gen.yaml missing", add)
	// 检查 gateway manifest 常见生成位置。
	if manifest := findGatewayManifest(root); manifest == "" {
		add("gateway manifest", CheckWarn, "gateway.manifest.json not found")
	} else {
		add("gateway manifest", CheckOK, manifest)
	}
	// 检查版本和 descriptor 文件。
	resolved, err := cfg.Resolve(root, path)
	if err != nil {
		add("bootstrap version", CheckFailed, err.Error())
	} else {
		add("bootstrap version", CheckOK, resolved.Version)
		if _, err = os.Stat(resolved.DescriptorFile); err != nil {
			add("descriptor file", CheckWarn, fmt.Sprintf("%s not found", resolved.DescriptorFile))
		} else {
			add("descriptor file", CheckOK, resolved.DescriptorFile)
		}
	}
	// 检查 S3 endpoint 格式。
	if cfg.S3.Endpoint == "" {
		add("s3 endpoint", CheckWarn, "s3.endpoint is empty")
	} else if _, err = url.ParseRequestURI(cfg.S3.Endpoint); err != nil {
		add("s3 endpoint", CheckWarn, err.Error())
	} else {
		add("s3 endpoint", CheckOK, cfg.S3.Endpoint)
	}
	// 检查 bucket。
	if cfg.S3.Bucket == "" {
		add("s3 bucket", CheckFailed, "s3.bucket is empty")
	} else {
		add("s3 bucket", CheckOK, cfg.S3.Bucket)
	}
	// 检查凭证来源是否可见。
	if hasCredentials(cfg.S3.Profile) {
		add("s3 credentials", CheckOK, "credential source detected")
	} else {
		add("s3 credentials", CheckWarn, "no env credentials, AWS profile, or ~/.aws/credentials detected")
	}
	return results, cfg, resolved
}

// templateVars 构造模板变量。
func (cfg *Config) templateVars(version string) map[string]string {
	return map[string]string{
		"service":   cfg.Service.Name,
		"app_id":    cfg.Service.AppID,
		"namespace": cfg.Service.Namespace,
		"module":    cfg.Service.Module,
		"version":   version,
		"endpoint":  strings.TrimRight(cfg.S3.Endpoint, "/"),
		"bucket":    cfg.S3.Bucket,
	}
}

// lookup 按点分路径从通用 map 中读取值。
func lookup(value any, path []string) (any, bool) {
	// 路径为空时返回当前值。
	if len(path) == 0 {
		return value, true
	}
	// 按 map 类型读取下一层。
	switch typed := value.(type) {
	case map[string]any:
		next, ok := typed[path[0]]
		if !ok {
			return nil, false
		}
		return lookup(next, path[1:])
	case map[any]any:
		next, ok := typed[path[0]]
		if !ok {
			return nil, false
		}
		return lookup(next, path[1:])
	default:
		return nil, false
	}
}

// expand 替换 ${name} 形式的模板变量。
func expand(input string, vars map[string]string) string {
	// 按变量表逐个替换。
	out := input
	for key, value := range vars {
		out = strings.ReplaceAll(out, "${"+key+"}", value)
	}
	return out
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	// 逐个 trim 后判断。
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// checkFile 检查文件是否存在。
func checkFile(root, name, okMessage, missingMessage string, add func(string, CheckStatus, string)) {
	// 缺失文件只提示 warn。
	if _, err := os.Stat(filepath.Join(root, name)); err != nil {
		add(name, CheckWarn, missingMessage)
		return
	}
	add(name, CheckOK, okMessage)
}

// checkAnyFile 检查任一候选文件是否存在。
func checkAnyFile(root, name string, candidates []string, okMessage, missingMessage string, add func(string, CheckStatus, string)) {
	// 任意候选路径存在即可通过。
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(root, candidate)); err == nil {
			add(name, CheckOK, okMessage)
			return
		}
	}
	add(name, CheckWarn, missingMessage)
}

// findGatewayManifest 查找 gateway.manifest.json 常见位置。
func findGatewayManifest(root string) string {
	// 仅检查本地路径，不连接 gateway。
	candidates := []string{
		filepath.Join(root, "dep", "protobuf", "gen", "gateway.manifest.json"),
		filepath.Join(root, "dep", "protobuf", "manifest", "gateway.manifest.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// hasCredentials 判断是否存在 AWS SDK 可用的凭证来源。
func hasCredentials(profile string) bool {
	// 环境变量中的 AK/SK 是最直接来源。
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	// profile 交给 AWS SDK 默认链处理，这里只判断是否声明。
	if profile != "" || os.Getenv("AWS_PROFILE") != "" {
		return true
	}
	// 本地 credentials 文件存在也视为有凭证来源。
	if home, err := os.UserHomeDir(); err == nil {
		if _, err = os.Stat(filepath.Join(home, ".aws", "credentials")); err == nil {
			return true
		}
	}
	return false
}
