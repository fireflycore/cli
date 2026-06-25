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

	// ProjectTypeService 表示业务服务项目。
	ProjectTypeService = "service"
	// ProjectTypeProto 表示 namespace proto 仓库项目。
	ProjectTypeProto = "proto"

	// DefaultLanguage 是默认项目语言。
	DefaultLanguage = "go"
	// DefaultNamespace 是默认服务命名空间。
	DefaultNamespace = "lhdht"
	// DefaultProtoSource 是 proto 项目默认 Buf build 来源。
	DefaultProtoSource = "."
	// DefaultProtoVersion 是 proto 项目默认 descriptor 发布版本。
	DefaultProtoVersion = "v0.0.1"
	// DefaultBootstrapFile 是默认启动配置路径。
	DefaultBootstrapFile = "conf/bootstrap.json"
	// DefaultBootstrapVersionPath 是默认版本字段路径，对应 bootstrapConf.app.version。
	DefaultBootstrapVersionPath = "app.version"
	// DefaultDescriptorDir 是 descriptor 默认输出目录。
	DefaultDescriptorDir = "dep/protobuf/gen"
	// DefaultDescriptorFileTemplate 是本地 descriptor 文件名模板。
	DefaultDescriptorFileTemplate = "${version}.pb"
	// DefaultProtoDescriptorFileTemplate 是 proto 项目本地版本化 descriptor 文件名模板。
	DefaultProtoDescriptorFileTemplate = "${namespace}/${version}.pb"
	// DefaultProtoCurrentFileTemplate 是 proto 项目本地 current descriptor 文件名模板。
	DefaultProtoCurrentFileTemplate = "${namespace}/current.pb"
	// DefaultObjectKeyTemplate 是 S3 对象 key 模板。
	DefaultObjectKeyTemplate = "${service}/${version}.pb"
	// DefaultProtoObjectKeyTemplate 是 proto 项目版本化 S3 对象 key 模板。
	DefaultProtoObjectKeyTemplate = "${namespace}/${version}.pb"
	// DefaultProtoCurrentObjectKeyTemplate 是 proto 项目 current S3 对象 key 模板。
	DefaultProtoCurrentObjectKeyTemplate = "${namespace}/current.pb"
	// DefaultDescriptorContentType 是 descriptor 上传时的默认 content type。
	DefaultDescriptorContentType = "application/octet-stream"
	// DefaultS3Region 是 S3 SDK 需要的默认 region。
	DefaultS3Region = "us-east-1"
	// DefaultS3Bucket 是 descriptor 主线使用的默认 bucket。
	DefaultS3Bucket = "descriptor"
	// DefaultDescriptorCurrentKeyTemplate 是 api-gateway watch 的 descriptor current key 模板。
	DefaultDescriptorCurrentKeyTemplate = "${namespace}/api-gateway/descriptor/current"
)

// Config 对应 .firefly/project.yaml 的完整结构。
type Config struct {
	// Schema 标识配置文件格式版本。
	Schema string `yaml:"schema" json:"schema"`
	// Project 保存项目类型。
	Project ProjectConfig `yaml:"project,omitempty" json:"project,omitempty"`
	// Service 保存服务身份信息。
	Service ServiceConfig `yaml:"service,omitempty" json:"service,omitempty"`
	// Proto 保存 proto 项目信息。
	Proto ProtoConfig `yaml:"proto,omitempty" json:"proto,omitempty"`
	// Bootstrap 保存服务版本读取规则。
	Bootstrap BootstrapConfig `yaml:"bootstrap,omitempty" json:"bootstrap,omitempty"`
	// Descriptor 保存 descriptor 本地和远端路径规则。
	Descriptor DescriptorConfig `yaml:"descriptor" json:"descriptor"`
	// Consul 保存 Consul KV 发布配置。
	Consul ConsulConfig `yaml:"consul,omitempty" json:"consul,omitempty"`
	// S3 保存 S3 兼容对象存储配置。
	S3 S3Config `yaml:"s3" json:"s3"`
}

// ProjectConfig 保存项目类型。
type ProjectConfig struct {
	// Type 是项目类型，支持 service 和 proto。
	Type string `yaml:"type" json:"type"`
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

// ProtoConfig 保存 proto 仓库维度的基础元信息。
type ProtoConfig struct {
	// Namespace 是 proto 仓库对应的 namespace。
	Namespace string `yaml:"namespace" json:"namespace"`
	// Repo 是 proto 仓库来源标识，例如 lhdht/backend/proto。
	Repo string `yaml:"repo,omitempty" json:"repo,omitempty"`
	// Module 是 Buf module 名或本地 module 标识。
	Module string `yaml:"module,omitempty" json:"module,omitempty"`
	// Source 是 buf build 的来源目录或 module。
	Source string `yaml:"source" json:"source"`
	// Version 是本次 descriptor 发布版本。
	Version string `yaml:"version" json:"version"`
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
	// CurrentFileTemplate 是本地 current descriptor 文件名模板。
	CurrentFileTemplate string `yaml:"current_file_template,omitempty" json:"current_file_template,omitempty"`
	// ObjectKeyTemplate 是 S3 object key 模板。
	ObjectKeyTemplate string `yaml:"object_key_template" json:"object_key_template"`
	// CurrentObjectKeyTemplate 是 S3 current object key 模板。
	CurrentObjectKeyTemplate string `yaml:"current_object_key_template,omitempty" json:"current_object_key_template,omitempty"`
	// RefTemplate 是 descriptor_ref URL 模板。
	RefTemplate string `yaml:"descriptor_ref_template,omitempty" json:"descriptor_ref_template,omitempty"`
	// Ref 是固定 descriptor_ref。
	Ref string `yaml:"descriptor_ref,omitempty" json:"descriptor_ref,omitempty"`
	// ContentType 是上传 descriptor 时使用的 content type。
	ContentType string `yaml:"content_type" json:"content_type"`
}

// ConsulConfig 保存 descriptor current 发布配置。
type ConsulConfig struct {
	// Address 是 Consul HTTP API 地址。
	Address string `yaml:"address,omitempty" json:"address,omitempty"`
	// DescriptorCurrentKey 是 api-gateway watch 的 descriptor current key。
	DescriptorCurrentKey string `yaml:"descriptor_current_key,omitempty" json:"descriptor_current_key,omitempty"`
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
	// ProjectType 是项目类型，支持 service 和 proto。
	ProjectType string
	// ServiceName 是服务名。
	ServiceName string
	// AppID 是 Firefly app id。
	AppID string
	// Namespace 是服务或 proto 项目命名空间。
	Namespace string
	// Language 是项目语言。
	Language string
	// Module 是 Go module 名。
	Module string
	// ProtoNamespace 是 proto 项目的 namespace。
	ProtoNamespace string
	// ProtoRepo 是 proto 仓库来源标识。
	ProtoRepo string
	// ProtoModule 是 Buf module 名或本地 module 标识。
	ProtoModule string
	// ProtoSource 是 buf build 的来源目录或 module。
	ProtoSource string
	// ProtoVersion 是 proto 项目 descriptor 发布版本。
	ProtoVersion string
	// BootstrapFile 是启动配置文件路径。
	BootstrapFile string
	// VersionPath 是版本字段路径。
	VersionPath string
	// DescriptorDir 是 descriptor 输出目录。
	DescriptorDir string
	// FileTemplate 是 descriptor 文件名模板。
	FileTemplate string
	// CurrentFileTemplate 是 current descriptor 文件名模板。
	CurrentFileTemplate string
	// ObjectKeyTemplate 是 S3 object key 模板。
	ObjectKeyTemplate string
	// CurrentObjectKeyTemplate 是 current S3 object key 模板。
	CurrentObjectKeyTemplate string
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
	// ConsulAddress 是 Consul HTTP API 地址。
	ConsulAddress string
	// DescriptorCurrentKey 是 api-gateway watch 的 descriptor current key。
	DescriptorCurrentKey string
	// Overwrite 控制是否覆盖已有 project.yaml。
	Overwrite bool
}

// Resolved 保存基于配置和版本推导出的发布信息。
type Resolved struct {
	// ConfigPath 是 project.yaml 路径。
	ConfigPath string
	// ProjectType 是项目类型。
	ProjectType string
	// Namespace 是服务或 proto 项目 namespace。
	Namespace string
	// Version 是从 bootstrapConf.app.version 或 proto.version 读取到的版本。
	Version string
	// DescriptorFile 是本地 descriptor 文件路径。
	DescriptorFile string
	// CurrentDescriptorFile 是本地 current descriptor 文件路径。
	CurrentDescriptorFile string
	// ObjectKey 是 S3 object key。
	ObjectKey string
	// CurrentObjectKey 是 S3 current object key。
	CurrentObjectKey string
	// DescriptorRef 是 gateway 可读取的 descriptor URL。
	DescriptorRef string
	// CurrentDescriptorRef 是 current descriptor URL。
	CurrentDescriptorRef string
	// DescriptorCurrentKey 是 api-gateway watch 的 Consul KV key。
	DescriptorCurrentKey string
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
	path := ConfigPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}
	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, path, err
	}
	cfg.ApplyDefaults()
	if cfg.Schema != SchemaVersion {
		return nil, path, fmt.Errorf("unsupported project schema %q", cfg.Schema)
	}
	if err = cfg.Validate(); err != nil {
		return nil, path, err
	}
	return &cfg, path, nil
}

// Save 写入项目配置。
func Save(root string, cfg *Config, overwrite bool) (string, error) {
	path := ConfigPath(root)
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return path, fmt.Errorf("%s already exists", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return path, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return path, err
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return path, err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return path, err
	}
	return path, os.WriteFile(path, data, 0644)
}

// NewConfig 根据命令输入构造项目配置。
func NewConfig(opts InitOptions) (*Config, error) {
	root := opts.Root
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	projectType := normalizeProjectType(opts.ProjectType)
	if projectType == "" {
		projectType = ProjectTypeService
	}
	if projectType != ProjectTypeService && projectType != ProjectTypeProto {
		return nil, fmt.Errorf("unsupported project type %q", opts.ProjectType)
	}

	module := strings.TrimSpace(opts.Module)
	if module == "" {
		if value, err := fsutil.ReadModule(root); err == nil {
			module = value
		}
	}

	cfg := &Config{
		Schema:  SchemaVersion,
		Project: ProjectConfig{Type: projectType},
		Descriptor: DescriptorConfig{
			Dir:         firstNonEmpty(opts.DescriptorDir, DefaultDescriptorDir),
			RefTemplate: strings.TrimSpace(opts.RefTemplate),
			Ref:         strings.TrimSpace(opts.DescriptorRef),
			ContentType: firstNonEmpty(opts.ContentType, DefaultDescriptorContentType),
		},
		S3: S3Config{
			Profile:        strings.TrimSpace(opts.S3Profile),
			Region:         firstNonEmpty(opts.S3Region, DefaultS3Region),
			Endpoint:       strings.TrimSpace(opts.S3Endpoint),
			Bucket:         firstNonEmpty(opts.S3Bucket, DefaultS3Bucket),
			ForcePathStyle: opts.ForcePathStyle,
		},
	}

	if projectType == ProjectTypeProto {
		cfg.Proto = ProtoConfig{
			Namespace: firstNonEmpty(opts.ProtoNamespace, opts.Namespace, DefaultNamespace),
			Repo:      strings.TrimSpace(opts.ProtoRepo),
			Module:    firstNonEmpty(opts.ProtoModule, module),
			Source:    firstNonEmpty(opts.ProtoSource, DefaultProtoSource),
			Version:   firstNonEmpty(opts.ProtoVersion, DefaultProtoVersion),
		}
		cfg.Descriptor.FileTemplate = firstNonEmpty(opts.FileTemplate, DefaultProtoDescriptorFileTemplate)
		cfg.Descriptor.CurrentFileTemplate = firstNonEmpty(opts.CurrentFileTemplate, DefaultProtoCurrentFileTemplate)
		cfg.Descriptor.ObjectKeyTemplate = firstNonEmpty(opts.ObjectKeyTemplate, DefaultProtoObjectKeyTemplate)
		cfg.Descriptor.CurrentObjectKeyTemplate = firstNonEmpty(opts.CurrentObjectKeyTemplate, DefaultProtoCurrentObjectKeyTemplate)
		cfg.Consul = ConsulConfig{
			Address:              strings.TrimSpace(opts.ConsulAddress),
			DescriptorCurrentKey: firstNonEmpty(opts.DescriptorCurrentKey, DefaultDescriptorCurrentKeyTemplate),
		}
	} else {
		service := firstNonEmpty(opts.ServiceName, filepath.Base(root))
		refTemplate := strings.TrimSpace(opts.RefTemplate)
		if refTemplate == "" && strings.TrimSpace(opts.DescriptorRef) == "" && strings.TrimSpace(opts.S3Endpoint) != "" {
			refTemplate = "${endpoint}/${bucket}/${service}/${version}.pb"
		}
		cfg.Service = ServiceConfig{
			Name:      service,
			AppID:     firstNonEmpty(opts.AppID, service),
			Namespace: firstNonEmpty(opts.Namespace, DefaultNamespace),
			Language:  strings.ToLower(firstNonEmpty(opts.Language, DefaultLanguage)),
			Module:    module,
		}
		cfg.Bootstrap = BootstrapConfig{
			File:        firstNonEmpty(opts.BootstrapFile, DefaultBootstrapFile),
			VersionPath: firstNonEmpty(opts.VersionPath, DefaultBootstrapVersionPath),
		}
		cfg.Descriptor.FileTemplate = firstNonEmpty(opts.FileTemplate, DefaultDescriptorFileTemplate)
		cfg.Descriptor.ObjectKeyTemplate = firstNonEmpty(opts.ObjectKeyTemplate, DefaultObjectKeyTemplate)
		cfg.Descriptor.RefTemplate = refTemplate
	}

	cfg.ApplyDefaults()
	return cfg, nil
}

// ApplyDefaults 为可省略字段补默认值。
func (cfg *Config) ApplyDefaults() {
	if cfg.Schema == "" {
		cfg.Schema = SchemaVersion
	}
	cfg.Project.Type = normalizeProjectType(cfg.Project.Type)
	if cfg.Project.Type == "" {
		cfg.Project.Type = ProjectTypeService
	}
	if cfg.Descriptor.Dir == "" {
		cfg.Descriptor.Dir = DefaultDescriptorDir
	}
	if cfg.IsProtoProject() {
		if cfg.Proto.Namespace == "" {
			cfg.Proto.Namespace = firstNonEmpty(cfg.Service.Namespace, DefaultNamespace)
		}
		if cfg.Proto.Source == "" {
			cfg.Proto.Source = DefaultProtoSource
		}
		if cfg.Proto.Version == "" {
			cfg.Proto.Version = DefaultProtoVersion
		}
		if cfg.Descriptor.FileTemplate == "" {
			cfg.Descriptor.FileTemplate = DefaultProtoDescriptorFileTemplate
		}
		if cfg.Descriptor.CurrentFileTemplate == "" {
			cfg.Descriptor.CurrentFileTemplate = DefaultProtoCurrentFileTemplate
		}
		if cfg.Descriptor.ObjectKeyTemplate == "" {
			cfg.Descriptor.ObjectKeyTemplate = DefaultProtoObjectKeyTemplate
		}
		if cfg.Descriptor.CurrentObjectKeyTemplate == "" {
			cfg.Descriptor.CurrentObjectKeyTemplate = DefaultProtoCurrentObjectKeyTemplate
		}
		if cfg.Consul.DescriptorCurrentKey == "" {
			cfg.Consul.DescriptorCurrentKey = DefaultDescriptorCurrentKeyTemplate
		}
	} else {
		if cfg.Service.Language == "" {
			cfg.Service.Language = DefaultLanguage
		}
		if cfg.Service.Namespace == "" {
			cfg.Service.Namespace = DefaultNamespace
		}
		if cfg.Service.AppID == "" {
			cfg.Service.AppID = cfg.Service.Name
		}
		if cfg.Bootstrap.File == "" {
			cfg.Bootstrap.File = DefaultBootstrapFile
		}
		if cfg.Bootstrap.VersionPath == "" {
			cfg.Bootstrap.VersionPath = DefaultBootstrapVersionPath
		}
		if cfg.Descriptor.FileTemplate == "" {
			cfg.Descriptor.FileTemplate = DefaultDescriptorFileTemplate
		}
		if cfg.Descriptor.ObjectKeyTemplate == "" {
			cfg.Descriptor.ObjectKeyTemplate = DefaultObjectKeyTemplate
		}
	}
	if cfg.Descriptor.ContentType == "" {
		cfg.Descriptor.ContentType = DefaultDescriptorContentType
	}
	if cfg.S3.Region == "" {
		cfg.S3.Region = DefaultS3Region
	}
	if cfg.S3.Bucket == "" {
		cfg.S3.Bucket = DefaultS3Bucket
	}
}

// Validate 校验 project.yaml 中 CLI 直接依赖的字段。
func (cfg *Config) Validate() error {
	switch cfg.Project.Type {
	case ProjectTypeService, ProjectTypeProto:
		return nil
	default:
		return fmt.Errorf("unsupported project type %q", cfg.Project.Type)
	}
}

// IsProtoProject 判断当前项目是否是 proto 仓库项目。
func (cfg *Config) IsProtoProject() bool {
	return cfg.Project.Type == ProjectTypeProto
}

// Resolve 推导本地 descriptor 文件和远端对象路径。
func (cfg *Config) Resolve(root, configPath string) (*Resolved, error) {
	return cfg.ResolveWithVersion(root, configPath, "")
}

// ResolveWithVersion 推导本地 descriptor 文件和远端对象路径，可覆盖 proto 版本。
func (cfg *Config) ResolveWithVersion(root, configPath, versionOverride string) (*Resolved, error) {
	var version string
	var err error
	if cfg.IsProtoProject() {
		version = firstNonEmpty(versionOverride, cfg.Proto.Version)
		if version == "" {
			return nil, fmt.Errorf("proto.version is empty")
		}
	} else {
		version, err = cfg.ReadVersion(root)
		if err != nil {
			return nil, err
		}
	}

	vars := cfg.templateVars(version)
	fileName := expand(cfg.Descriptor.FileTemplate, vars)
	objectKey := filepath.ToSlash(expand(cfg.Descriptor.ObjectKeyTemplate, vars))
	resolved := &Resolved{
		ConfigPath:     configPath,
		ProjectType:    cfg.Project.Type,
		Namespace:      cfg.namespace(),
		Version:        version,
		DescriptorFile: filepath.Join(root, cfg.Descriptor.Dir, fileName),
		ObjectKey:      objectKey,
		DescriptorRef:  cfg.ResolveDescriptorRef(version, objectKey),
	}
	if cfg.IsProtoProject() {
		currentFileName := expand(cfg.Descriptor.CurrentFileTemplate, vars)
		currentObjectKey := filepath.ToSlash(expand(cfg.Descriptor.CurrentObjectKeyTemplate, vars))
		resolved.CurrentDescriptorFile = filepath.Join(root, cfg.Descriptor.Dir, currentFileName)
		resolved.CurrentObjectKey = currentObjectKey
		resolved.CurrentDescriptorRef = cfg.ResolveDescriptorRef(version, currentObjectKey)
		resolved.DescriptorCurrentKey = filepath.ToSlash(expand(cfg.Consul.DescriptorCurrentKey, vars))
	}
	return resolved, nil
}

// ReadVersion 从 bootstrap 配置读取服务版本。
func (cfg *Config) ReadVersion(root string) (string, error) {
	path := filepath.Join(root, cfg.Bootstrap.File)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read bootstrap config: %w", err)
	}
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
	vars := cfg.templateVars(version)
	vars["key"] = objectKey
	if cfg.Descriptor.RefTemplate != "" {
		return expand(cfg.Descriptor.RefTemplate, vars)
	}
	if cfg.Descriptor.Ref != "" {
		return expand(cfg.Descriptor.Ref, vars)
	}
	if cfg.IsProtoProject() && cfg.S3.Bucket != "" && objectKey != "" {
		return "s3://" + cfg.S3.Bucket + "/" + objectKey
	}
	if cfg.S3.Endpoint == "" || cfg.S3.Bucket == "" || objectKey == "" {
		return ""
	}
	return strings.TrimRight(cfg.S3.Endpoint, "/") + "/" + cfg.S3.Bucket + "/" + objectKey
}

// ReadFireflyModules 从 go.mod 中提取 Firefly 依赖版本。
func ReadFireflyModules(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil
	}
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
	results := make([]CheckResult, 0, 12)
	add := func(name string, status CheckStatus, message string) {
		results = append(results, CheckResult{Name: name, Status: status, Message: message})
	}

	cfg, path, err := Load(root)
	if err != nil {
		add("project config", CheckFailed, err.Error())
		return results, nil, nil
	}
	add("project config", CheckOK, path)
	add("project type", CheckOK, cfg.Project.Type)

	if module, err := fsutil.ReadModule(root); err != nil {
		add("go.mod", CheckWarn, "go.mod not found or module line missing")
	} else {
		add("go.mod", CheckOK, module)
	}
	if !cfg.IsProtoProject() {
		checkAnyFile(root, "makefile", []string{"Makefile", "makefile"}, "makefile exists", "Makefile or makefile missing", add)
	}
	checkFile(root, "buf.yaml", "buf.yaml exists", "buf.yaml missing", add)
	checkFile(root, "buf.gen.yaml", "buf.gen.yaml exists", "buf.gen.yaml missing", add)
	if !cfg.IsProtoProject() {
		if manifest := findGatewayManifest(root); manifest == "" {
			add("gateway manifest", CheckWarn, "gateway.manifest.json not found")
		} else {
			add("gateway manifest", CheckOK, manifest)
		}
	}

	resolved, err := cfg.Resolve(root, path)
	if err != nil {
		if cfg.IsProtoProject() {
			add("proto version", CheckFailed, err.Error())
		} else {
			add("bootstrap version", CheckFailed, err.Error())
		}
	} else {
		if cfg.IsProtoProject() {
			add("proto version", CheckOK, resolved.Version)
		} else {
			add("bootstrap version", CheckOK, resolved.Version)
		}
		if _, err = os.Stat(resolved.DescriptorFile); err != nil {
			add("descriptor file", CheckWarn, fmt.Sprintf("%s not found", resolved.DescriptorFile))
		} else {
			add("descriptor file", CheckOK, resolved.DescriptorFile)
		}
		if cfg.IsProtoProject() {
			if _, err = os.Stat(resolved.CurrentDescriptorFile); err != nil {
				add("current descriptor file", CheckWarn, fmt.Sprintf("%s not found", resolved.CurrentDescriptorFile))
			} else {
				add("current descriptor file", CheckOK, resolved.CurrentDescriptorFile)
			}
			if resolved.DescriptorCurrentKey == "" {
				add("descriptor current key", CheckFailed, "consul.descriptor_current_key is empty")
			} else {
				add("descriptor current key", CheckOK, resolved.DescriptorCurrentKey)
			}
		}
	}

	if cfg.S3.Endpoint == "" {
		add("s3 endpoint", CheckWarn, "s3.endpoint is empty")
	} else if _, err = url.ParseRequestURI(cfg.S3.Endpoint); err != nil {
		add("s3 endpoint", CheckWarn, err.Error())
	} else {
		add("s3 endpoint", CheckOK, cfg.S3.Endpoint)
	}
	if cfg.S3.Bucket == "" {
		add("s3 bucket", CheckFailed, "s3.bucket is empty")
	} else {
		add("s3 bucket", CheckOK, cfg.S3.Bucket)
	}
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
		"service":    cfg.Service.Name,
		"app_id":     cfg.Service.AppID,
		"namespace":  cfg.namespace(),
		"module":     firstNonEmpty(cfg.Proto.Module, cfg.Service.Module),
		"repo":       cfg.Proto.Repo,
		"proto_repo": cfg.Proto.Repo,
		"source":     cfg.Proto.Source,
		"version":    version,
		"endpoint":   strings.TrimRight(cfg.S3.Endpoint, "/"),
		"bucket":     cfg.S3.Bucket,
	}
}

func (cfg *Config) namespace() string {
	if cfg.IsProtoProject() {
		return cfg.Proto.Namespace
	}
	return cfg.Service.Namespace
}

func normalizeProjectType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
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
