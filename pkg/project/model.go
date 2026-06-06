package project

const (
	// ConfigDirName 是项目元信息目录名。
	ConfigDirName = ".firefly"
	// ConfigFileName 是项目元信息文件名。
	ConfigFileName = "project.yaml"
	// SchemaVersion 是当前 project.yaml 的 schema 版本。
	SchemaVersion = "firefly.project.v1"

	// DefaultLanguage 是默认项目语言。
	DefaultLanguage = "go"
	// DefaultNamespace 是默认服务命名空间。
	DefaultNamespace = "lhdht"
	// DefaultBootstrapFile 是默认启动配置路径。
	DefaultBootstrapFile = "conf/bootstrap.json"
	// DefaultBootstrapVersion 是默认版本字段路径，对应 bootstrapConf.app.version。
	DefaultBootstrapVersion = "app.version"
	// DefaultDescriptorDir 是默认 descriptor 输出目录。
	DefaultDescriptorDir = "dist/descriptors"
	// DefaultFileTemplate 是默认 descriptor 文件名模板。
	DefaultFileTemplate = "${version}.pb"
	// DefaultObjectKeyTemplate 是默认 S3 对象 key 模板。
	DefaultObjectKeyTemplate = "${service}/${version}.pb"
	// DefaultContentType 是 descriptor 上传时的默认 content type。
	DefaultContentType = "application/octet-stream"
	// DefaultS3Region 是 S3 SDK 要求的默认 region。
	DefaultS3Region = "us-east-1"
	// DefaultS3Bucket 是 descriptor 主线使用的默认 bucket。
	DefaultS3Bucket = "descriptor"
)

// Config 对应 .firefly/project.yaml 的完整结构。
type Config struct {
	// Schema 标记配置 schema，避免未来格式变更时误读。
	Schema string `yaml:"schema" json:"schema"`
	// Service 保存当前业务服务身份信息。
	Service ServiceConfig `yaml:"service" json:"service"`
	// Bootstrap 保存启动配置文件和版本字段路径。
	Bootstrap BootstrapConfig `yaml:"bootstrap" json:"bootstrap"`
	// Descriptor 保存 descriptor 本地文件和对象存储路径规则。
	Descriptor DescriptorConfig `yaml:"descriptor" json:"descriptor"`
	// S3 保存 S3 兼容对象存储配置。
	S3 S3Config `yaml:"s3" json:"s3"`
}

// ServiceConfig 保存服务维度的基础元信息。
type ServiceConfig struct {
	// Name 是服务名，也用于默认对象 key。
	Name string `yaml:"name" json:"name"`
	// AppID 是 Firefly 应用 ID。
	AppID string `yaml:"app_id" json:"app_id"`
	// Namespace 是服务命名空间。
	Namespace string `yaml:"namespace" json:"namespace"`
	// Language 是项目语言。
	Language string `yaml:"language" json:"language"`
	// Module 是 Go module 名。
	Module string `yaml:"module" json:"module"`
}

// BootstrapConfig 描述如何从启动配置中读取服务版本。
type BootstrapConfig struct {
	// File 是启动配置文件路径。
	File string `yaml:"file" json:"file"`
	// VersionPath 是版本字段路径，默认 app.version。
	VersionPath string `yaml:"version_path" json:"version_path"`
}

// DescriptorConfig 描述 descriptor 本地文件和发布 URL 的推导规则。
type DescriptorConfig struct {
	// Dir 是本地 descriptor 输出目录。
	Dir string `yaml:"dir" json:"dir"`
	// FileTemplate 是本地 descriptor 文件名模板。
	FileTemplate string `yaml:"file_template" json:"file_template"`
	// ObjectKeyTemplate 是 S3 对象 key 模板。
	ObjectKeyTemplate string `yaml:"object_key_template" json:"object_key_template"`
	// RefTemplate 是 descriptor_ref 的 URL 模板。
	RefTemplate string `yaml:"descriptor_ref_template,omitempty" json:"descriptor_ref_template,omitempty"`
	// Ref 是可选固定 descriptor_ref。
	Ref string `yaml:"descriptor_ref,omitempty" json:"descriptor_ref,omitempty"`
	// ContentType 是上传对象的 content type。
	ContentType string `yaml:"content_type" json:"content_type"`
}

// S3Config 保存 S3 兼容对象存储配置。
type S3Config struct {
	// Profile 是 AWS 共享配置 profile。
	Profile string `yaml:"profile" json:"profile"`
	// Region 是 S3 region。
	Region string `yaml:"region" json:"region"`
	// Endpoint 是 S3 兼容 endpoint。
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// Bucket 是 descriptor 目标 bucket。
	Bucket string `yaml:"bucket" json:"bucket"`
	// ForcePathStyle 控制是否使用路径风格访问。
	ForcePathStyle bool `yaml:"force_path_style" json:"force_path_style"`
}

// Resolved 是根据 project.yaml 和 bootstrap 版本推导出的运行时结果。
type Resolved struct {
	// Root 是项目根目录。
	Root string
	// ConfigPath 是 project.yaml 路径。
	ConfigPath string
	// Version 是从 bootstrapConf.app.version 读取到的服务版本。
	Version string
	// DescriptorFile 是本地 descriptor 文件完整路径。
	DescriptorFile string
	// ObjectKey 是 S3 对象 key。
	ObjectKey string
	// DescriptorRef 是 api-gateway 可访问的 descriptor_ref。
	DescriptorRef string
}

// InitOptions 是 project init 命令传入的初始化参数。
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
	// BootstrapFile 是启动配置路径。
	BootstrapFile string
	// VersionPath 是版本字段路径。
	VersionPath string
	// DescriptorDir 是 descriptor 输出目录。
	DescriptorDir string
	// FileTemplate 是 descriptor 文件名模板。
	FileTemplate string
	// ObjectKeyTemplate 是 S3 对象 key 模板。
	ObjectKeyTemplate string
	// DescriptorRef 是固定 descriptor_ref。
	DescriptorRef string
	// RefTemplate 是 descriptor_ref 模板。
	RefTemplate string
	// ContentType 是上传 content type。
	ContentType string
	// S3Profile 是 AWS 共享配置 profile。
	S3Profile string
	// S3Region 是 S3 region。
	S3Region string
	// S3Endpoint 是 S3 兼容 endpoint。
	S3Endpoint string
	// S3Bucket 是目标 bucket。
	S3Bucket string
	// ForcePathStyle 控制是否使用路径风格访问。
	ForcePathStyle bool
	// Overwrite 控制是否覆盖已有 project.yaml。
	Overwrite bool
}

// CheckStatus 是 project check 的单项检查状态。
type CheckStatus string

const (
	// CheckOK 表示检查通过。
	CheckOK CheckStatus = "ok"
	// CheckWarn 表示检查有提醒但不阻断命令。
	CheckWarn CheckStatus = "warn"
	// CheckFailed 表示检查失败并应导致命令返回错误。
	CheckFailed CheckStatus = "failed"
)

// CheckResult 表示 project check 的一条检查结果。
type CheckResult struct {
	// Name 是检查项名称。
	Name string
	// Status 是检查项状态。
	Status CheckStatus
	// Message 是检查结果说明。
	Message string
}
