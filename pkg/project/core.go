package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// moduleLineRegexp 用于从 go.mod 中提取 module 名。
var moduleLineRegexp = regexp.MustCompile(`^\s*module\s+(\S+)\s*$`)

// ConfigPath 返回指定项目根目录下的 .firefly/project.yaml 路径。
func ConfigPath(root string) string {
	// 使用 filepath.Join 兼容不同操作系统路径分隔符。
	return filepath.Join(root, ConfigDirName, ConfigFileName)
}

// Exists 判断指定项目根目录是否已经存在 project.yaml。
func Exists(root string) bool {
	// 尝试 stat 配置文件。
	_, err := os.Stat(ConfigPath(root))
	// stat 成功表示文件存在。
	return err == nil
}

// Load 读取并解析 .firefly/project.yaml。
func Load(root string) (*Config, string, error) {
	// 计算配置文件路径。
	path := ConfigPath(root)
	// 读取配置文件内容。
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, path, err
	}

	// 创建配置结构体承载 YAML 内容。
	var cfg Config
	// 将 YAML 反序列化为 Config。
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, path, err
	}
	// 补齐可省略的默认值。
	cfg.ApplyDefaults()

	// 校验 schema，避免误读其他版本配置。
	if cfg.Schema != SchemaVersion {
		return nil, path, fmt.Errorf("unsupported project schema %q", cfg.Schema)
	}

	// 返回配置对象和配置文件路径。
	return &cfg, path, nil
}

// Save 将项目配置写入 .firefly/project.yaml。
func Save(root string, cfg *Config, overwrite bool) (string, error) {
	// 计算目标配置文件路径。
	path := ConfigPath(root)
	// 如果不允许覆盖，则先检查文件是否已经存在。
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

	// 写入前再次补齐默认值。
	cfg.ApplyDefaults()
	// 将配置序列化为 YAML。
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return path, err
	}

	// 写入配置文件。
	return path, os.WriteFile(path, data, 0644)
}

// NewConfig 根据 project init 的参数生成默认项目配置。
func NewConfig(opts InitOptions) (*Config, error) {
	// 优先使用传入根目录。
	root := opts.Root
	// 未传入根目录时使用当前工作目录。
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// 读取显式服务名。
	service := strings.TrimSpace(opts.ServiceName)
	// 未传服务名时使用当前目录名。
	if service == "" {
		service = filepath.Base(root)
	}

	// 读取显式 module。
	module := strings.TrimSpace(opts.Module)
	// 未传 module 时尝试从 go.mod 中读取。
	if module == "" {
		module = ReadModule(root)
	}

	// 读取显式 app id。
	appID := strings.TrimSpace(opts.AppID)
	// 未传 app id 时默认等于服务名。
	if appID == "" {
		appID = service
	}

	// 读取 descriptor_ref 模板。
	refTemplate := opts.RefTemplate
	// 如果用户只给了 endpoint，则自动生成路径风格 URL 模板。
	if refTemplate == "" && opts.DescriptorRef == "" && opts.S3Endpoint != "" {
		refTemplate = "${endpoint}/${bucket}/${service}/${version}.pb"
	}

	// 组装完整配置结构。
	cfg := &Config{
		// 写入固定 schema 版本。
		Schema: SchemaVersion,
		// 写入服务元信息。
		Service: ServiceConfig{
			Name:      service,
			AppID:     appID,
			Namespace: firstNonEmpty(opts.Namespace, DefaultNamespace),
			Language:  firstNonEmpty(strings.ToLower(opts.Language), DefaultLanguage),
			Module:    module,
		},
		// 写入 bootstrap 读取规则。
		Bootstrap: BootstrapConfig{
			File:        firstNonEmpty(opts.BootstrapFile, DefaultBootstrapFile),
			VersionPath: firstNonEmpty(opts.VersionPath, DefaultBootstrapVersion),
		},
		// 写入 descriptor 路径规则。
		Descriptor: DescriptorConfig{
			Dir:               firstNonEmpty(opts.DescriptorDir, DefaultDescriptorDir),
			FileTemplate:      firstNonEmpty(opts.FileTemplate, DefaultFileTemplate),
			ObjectKeyTemplate: firstNonEmpty(opts.ObjectKeyTemplate, DefaultObjectKeyTemplate),
			RefTemplate:       refTemplate,
			Ref:               opts.DescriptorRef,
			ContentType:       firstNonEmpty(opts.ContentType, DefaultContentType),
		},
		// 写入 S3 兼容对象存储配置。
		S3: S3Config{
			Profile:        opts.S3Profile,
			Region:         firstNonEmpty(opts.S3Region, DefaultS3Region),
			Endpoint:       opts.S3Endpoint,
			Bucket:         firstNonEmpty(opts.S3Bucket, DefaultS3Bucket),
			ForcePathStyle: opts.ForcePathStyle,
		},
	}
	// 再次补齐空字段，保证调用方传空值也能得到完整配置。
	cfg.ApplyDefaults()
	return cfg, nil
}

// ApplyDefaults 为可省略字段补齐默认值。
func (cfg *Config) ApplyDefaults() {
	// schema 为空时写入当前 schema。
	if cfg.Schema == "" {
		cfg.Schema = SchemaVersion
	}
	// 语言为空时默认 go。
	if cfg.Service.Language == "" {
		cfg.Service.Language = DefaultLanguage
	}
	// 命名空间为空时默认 lhdht。
	if cfg.Service.Namespace == "" {
		cfg.Service.Namespace = DefaultNamespace
	}
	// app id 为空时默认等于服务名。
	if cfg.Service.AppID == "" {
		cfg.Service.AppID = cfg.Service.Name
	}
	// bootstrap 文件为空时默认 conf/bootstrap.json。
	if cfg.Bootstrap.File == "" {
		cfg.Bootstrap.File = DefaultBootstrapFile
	}
	// 版本字段为空时默认 app.version。
	if cfg.Bootstrap.VersionPath == "" {
		cfg.Bootstrap.VersionPath = DefaultBootstrapVersion
	}
	// descriptor 目录为空时默认 dist/descriptors。
	if cfg.Descriptor.Dir == "" {
		cfg.Descriptor.Dir = DefaultDescriptorDir
	}
	// 文件名模板为空时默认 ${version}.pb。
	if cfg.Descriptor.FileTemplate == "" {
		cfg.Descriptor.FileTemplate = DefaultFileTemplate
	}
	// 对象 key 模板为空时默认 ${service}/${version}.pb。
	if cfg.Descriptor.ObjectKeyTemplate == "" {
		cfg.Descriptor.ObjectKeyTemplate = DefaultObjectKeyTemplate
	}
	// content type 为空时默认二进制流。
	if cfg.Descriptor.ContentType == "" {
		cfg.Descriptor.ContentType = DefaultContentType
	}
	// region 为空时使用 SDK 可接受的默认 region。
	if cfg.S3.Region == "" {
		cfg.S3.Region = DefaultS3Region
	}
	// bucket 为空时默认 descriptor。
	if cfg.S3.Bucket == "" {
		cfg.S3.Bucket = DefaultS3Bucket
	}
}

// Resolve 根据 project.yaml 和 bootstrapConf.app.version 推导 descriptor 发布信息。
func (cfg *Config) Resolve(root, configPath string) (*Resolved, error) {
	// 从 bootstrap 配置中读取服务版本。
	version, err := cfg.ReadVersion(root)
	if err != nil {
		return nil, err
	}

	// 构造模板变量。
	vars := cfg.templateVars(version)
	// 展开本地 descriptor 文件名。
	fileName := expand(cfg.Descriptor.FileTemplate, vars)
	// 展开 S3 对象 key，并统一为 URL 风格斜杠。
	objectKey := filepath.ToSlash(expand(cfg.Descriptor.ObjectKeyTemplate, vars))
	// 根据模板或 endpoint 推导 descriptor_ref。
	descriptorRef := cfg.ResolveDescriptorRef(version, objectKey)

	// 返回解析结果。
	return &Resolved{
		Root:           root,
		ConfigPath:     configPath,
		Version:        version,
		DescriptorFile: filepath.Join(root, cfg.Descriptor.Dir, fileName),
		ObjectKey:      objectKey,
		DescriptorRef:  descriptorRef,
	}, nil
}

// ReadVersion 从启动配置中读取服务版本。
func (cfg *Config) ReadVersion(root string) (string, error) {
	// 计算 bootstrap 文件路径。
	path := filepath.Join(root, cfg.Bootstrap.File)
	// 读取 bootstrap 文件。
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read bootstrap config: %w", err)
	}

	// raw 保存 JSON/YAML 解析后的通用结构。
	var raw any
	// 根据文件扩展名选择 YAML 或 JSON 解析。
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &raw)
	default:
		err = json.Unmarshal(data, &raw)
	}
	if err != nil {
		return "", fmt.Errorf("parse bootstrap config: %w", err)
	}

	// 按 app.version 这类路径读取值。
	value, ok := lookup(raw, strings.Split(cfg.Bootstrap.VersionPath, "."))
	if !ok {
		return "", fmt.Errorf("bootstrap version path %q not found", cfg.Bootstrap.VersionPath)
	}

	// 将值转成字符串，并去掉首尾空白。
	version := strings.TrimSpace(fmt.Sprint(value))
	// 空版本是无效配置。
	if version == "" || version == "<nil>" {
		return "", fmt.Errorf("bootstrap version path %q is empty", cfg.Bootstrap.VersionPath)
	}
	return version, nil
}

// ResolveDescriptorRef 推导 api-gateway 可访问的 descriptor_ref。
func (cfg *Config) ResolveDescriptorRef(version, objectKey string) string {
	// 构造模板变量。
	vars := cfg.templateVars(version)
	// key 表示完整对象 key。
	vars["key"] = objectKey

	// 优先使用 descriptor_ref_template。
	if cfg.Descriptor.RefTemplate != "" {
		return expand(cfg.Descriptor.RefTemplate, vars)
	}
	// 其次使用固定 descriptor_ref，并允许其中带模板变量。
	if cfg.Descriptor.Ref != "" {
		return expand(cfg.Descriptor.Ref, vars)
	}
	// 缺少 endpoint、bucket 或 key 时无法自动推导 URL。
	if cfg.S3.Endpoint == "" || cfg.S3.Bucket == "" || objectKey == "" {
		return ""
	}

	// 默认按路径风格 URL 拼出 descriptor_ref。
	return strings.TrimRight(cfg.S3.Endpoint, "/") + "/" + cfg.S3.Bucket + "/" + objectKey
}

// templateVars 构造 descriptor 模板展开所需变量。
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

// ReadModule 从 go.mod 中读取 module 名。
func ReadModule(root string) string {
	// 读取 go.mod。
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	// 逐行寻找 module 声明。
	for _, line := range strings.Split(string(data), "\n") {
		matches := moduleLineRegexp.FindStringSubmatch(line)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

// ReadFireflyModules 从 go.mod 中提取 github.com/fireflycore/* 依赖版本。
func ReadFireflyModules(root string) map[string]string {
	// 读取 go.mod。
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil
	}

	// deps 保存 fireflycore 依赖名和版本。
	deps := make(map[string]string)
	// 逐行扫描 require 内容。
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

// Check 执行不连接运行时组件的本地项目检查。
func Check(root string) ([]CheckResult, *Config, *Resolved) {
	// results 收集所有检查结果。
	results := make([]CheckResult, 0, 10)
	// add 用于统一追加检查项。
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

	// 检查 go.mod 是否可识别。
	if ReadModule(root) == "" {
		add("go.mod", CheckWarn, "go.mod not found or module line missing")
	} else {
		add("go.mod", CheckOK, "module detected")
	}

	// 检查 Makefile 是否存在，只警告不失败。
	checkFile(root, "Makefile", "Makefile exists", "Makefile missing", add)
	// 检查 buf.yaml 是否存在。
	checkFile(root, "buf.yaml", "buf.yaml exists", "buf.yaml missing", add)
	// 检查 buf.gen.yaml 是否存在。
	checkFile(root, "buf.gen.yaml", "buf.gen.yaml exists", "buf.gen.yaml missing", add)

	// 检查 gateway.manifest.json 的常见位置。
	if manifest := findGatewayManifest(root); manifest == "" {
		add("gateway manifest", CheckWarn, "gateway.manifest.json not found")
	} else {
		add("gateway manifest", CheckOK, manifest)
	}

	// 推导版本、descriptor 文件和对象 key。
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

	// 检查 S3 endpoint 是否为空或格式异常。
	if cfg.S3.Endpoint == "" {
		add("s3 endpoint", CheckWarn, "s3.endpoint is empty")
	} else if _, err = url.ParseRequestURI(cfg.S3.Endpoint); err != nil {
		add("s3 endpoint", CheckWarn, err.Error())
	} else {
		add("s3 endpoint", CheckOK, cfg.S3.Endpoint)
	}

	// bucket 为空会阻断 descriptor push。
	if cfg.S3.Bucket == "" {
		add("s3 bucket", CheckFailed, "s3.bucket is empty")
	} else {
		add("s3 bucket", CheckOK, cfg.S3.Bucket)
	}

	// 检查是否能找到某种凭证来源。
	if hasCredentials(cfg.S3.Profile) {
		add("s3 credentials", CheckOK, "credential source detected")
	} else {
		add("s3 credentials", CheckWarn, "no env credentials, AWS profile, or ~/.aws/credentials detected")
	}

	return results, cfg, resolved
}

// findGatewayManifest 查找 gateway.manifest.json 的常见生成位置。
func findGatewayManifest(root string) string {
	// 按当前文档和历史产物位置列出候选路径。
	candidates := []string{
		filepath.Join(root, "dep", "protobuf", "gen", "gateway.manifest.json"),
		filepath.Join(root, "dep", "protobuf", "manifest", "gateway.manifest.json"),
	}
	// 找到第一个存在的候选文件就返回。
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// checkFile 检查某个项目根目录下的文件是否存在。
func checkFile(root, name, okMessage, missingMessage string, add func(string, CheckStatus, string)) {
	// 文件不存在时追加 warn。
	if _, err := os.Stat(filepath.Join(root, name)); err != nil {
		add(name, CheckWarn, missingMessage)
		return
	}
	// 文件存在时追加 ok。
	add(name, CheckOK, okMessage)
}

// hasCredentials 检查是否存在可供 AWS SDK 使用的凭证来源。
func hasCredentials(profile string) bool {
	// 显式环境变量是最直接的凭证来源。
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	// profile 由 AWS SDK 默认链处理，这里只做存在性判断。
	if profile != "" || os.Getenv("AWS_PROFILE") != "" {
		return true
	}
	// ~/.aws/credentials 存在时也视为有凭证来源。
	if home, err := os.UserHomeDir(); err == nil {
		if _, err = os.Stat(filepath.Join(home, ".aws", "credentials")); err == nil {
			return true
		}
	}
	return false
}

// lookup 按点分路径递归读取 map 中的值。
func lookup(value any, path []string) (any, bool) {
	// 路径耗尽时返回当前值。
	if len(path) == 0 {
		return value, true
	}
	// 根据 map 类型读取下一层。
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

// expand 使用 ${name} 形式的变量替换模板字符串。
func expand(input string, vars map[string]string) string {
	// out 保存逐步替换后的字符串。
	out := input
	// 遍历所有变量并执行全量替换。
	for key, value := range vars {
		out = strings.ReplaceAll(out, "${"+key+"}", value)
	}
	return out
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	// 逐个检查候选值。
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
