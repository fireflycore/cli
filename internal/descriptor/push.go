package descriptor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fireflycore/cli/internal/project"
)

const descriptorCurrentSchema = "firefly.api_gateway.descriptor.v1"

// BuildOptions 表示 descriptor build 命令传入的参数。
type BuildOptions struct {
	// Root 是 proto 项目根目录。
	Root string
	// Version 覆盖 project.yaml 中的 proto.version。
	Version string
	// Source 覆盖 project.yaml 中的 proto.source。
	Source string
	// Out 覆盖版本化 descriptor 输出文件。
	Out string
	// Buf 是 Buf CLI 路径，默认 buf。
	Buf string
}

// BuildResult 表示 descriptor build 的结果。
type BuildResult struct {
	// ConfigPath 是 project.yaml 路径。
	ConfigPath string
	// Namespace 是 proto 项目 namespace。
	Namespace string
	// Version 是本次 descriptor 版本。
	Version string
	// Source 是 buf build 来源。
	Source string
	// File 是版本化 descriptor 文件。
	File string
	// CurrentFile 是本地 current descriptor 文件。
	CurrentFile string
	// SHA256 是 descriptor 文件 sha256 摘要。
	SHA256 string
	// Size 是 descriptor 文件字节数。
	Size int64
}

// PushOptions 表示 descriptor push 命令传入的参数。
type PushOptions struct {
	// Root 是业务服务或 proto 项目根目录。
	Root string
	// Version 覆盖 proto 项目中的 proto.version。
	Version string
	// File 显式覆盖本地 descriptor 文件路径。
	File string
	// CurrentFile 显式覆盖本地 current descriptor 文件路径。
	CurrentFile string
	// Bucket 显式覆盖目标 bucket。
	Bucket string
	// Key 显式覆盖目标 object key。
	Key string
	// CurrentKey 显式覆盖 current object key。
	CurrentKey string
	// Endpoint 显式覆盖 S3 兼容 endpoint。
	Endpoint string
	// Region 显式覆盖 S3 region。
	Region string
	// Profile 指定 AWS shared config profile。
	Profile string
	// ForcePathStyle 控制是否使用 path-style 地址。
	ForcePathStyle bool
	// ContentType 覆盖上传对象 content type。
	ContentType string
	// VersionedRef 覆盖命令最终打印的 versioned descriptor ref。
	VersionedRef string
	// CurrentRef 覆盖 current descriptor ref。
	CurrentRef string
	// AccessKeyID 是显式传入的 access key id。
	AccessKeyID string
	// SecretAccessKey 是显式传入的 secret access key。
	SecretAccessKey string
	// SessionToken 是 STS 临时凭证 token。
	SessionToken string
	// SkipCurrentObject 表示 proto 项目只上传版本化对象。
	SkipCurrentObject bool
	// DryRun 表示只解析和计算摘要，不上传对象。
	DryRun bool
}

// PushResult 表示 descriptor push 的解析和上传结果。
type PushResult struct {
	// ProjectType 是项目类型。
	ProjectType string
	// Namespace 是 proto 项目 namespace。
	Namespace string
	// Version 是 descriptor 版本。
	Version string
	// File 是最终使用的本地 descriptor 文件。
	File string
	// CurrentFile 是最终使用的本地 current descriptor 文件。
	CurrentFile string
	// Bucket 是最终使用的目标 bucket。
	Bucket string
	// Key 是最终使用的 object key。
	Key string
	// CurrentKey 是 current object key。
	CurrentKey string
	// VersionedRef 是最终输出的 versioned descriptor ref。
	VersionedRef string
	// CurrentRef 是 current descriptor ref。
	CurrentRef string
	// SHA256 是 descriptor 文件 sha256 摘要。
	SHA256 string
	// Size 是 descriptor 文件字节数。
	Size int64
	// DryRun 表示本次没有真实上传。
	DryRun bool
	// PushedCurrent 表示本次上传了 current 对象。
	PushedCurrent bool
}

// PublishOptions 表示 descriptor publish 命令传入的参数。
type PublishOptions struct {
	PushOptions
	// Source 覆盖 project.yaml 中的 proto.source。
	Source string
	// Out 覆盖版本化 descriptor 输出文件。
	Out string
	// Buf 是 Buf CLI 路径，默认 buf。
	Buf string
	// SkipBuild 表示使用已有本地 pb。
	SkipBuild bool
	// SkipConsul 表示不更新 Consul KV。
	SkipConsul bool
	// ConsulAddress 显式覆盖 Consul HTTP API 地址。
	ConsulAddress string
	// SourceRevision 指定 descriptor current JSON 的 source_revision。
	SourceRevision string
}

// PublishResult 表示 descriptor publish 的结果。
type PublishResult struct {
	// Build 是 build 阶段结果，SkipBuild 时为空。
	Build *BuildResult
	// Push 是 push 阶段结果。
	Push *PushResult
	// CurrentKey 是 Consul descriptor current key。
	CurrentKey string
	// ConsulAddress 是 Consul HTTP API 地址。
	ConsulAddress string
	// CurrentJSON 是写入 Consul 的 descriptor current JSON。
	CurrentJSON []byte
	// DryRun 表示没有真实上传对象或写 Consul。
	DryRun bool
	// PublishedAt 是生成 current JSON 的时间。
	PublishedAt time.Time
}

type descriptorCurrentDocument struct {
	Schema         string `json:"schema"`
	Namespace      string `json:"namespace"`
	Version        string `json:"version"`
	Ref            string `json:"ref"`
	CurrentRef     string `json:"current_ref"`
	SHA256         string `json:"sha256"`
	ProtoRepo      string `json:"proto_repo,omitempty"`
	SourceRevision string `json:"source_revision,omitempty"`
	PublishedAt    string `json:"published_at"`
}

// Build 调用 Buf CLI 生成 proto 项目的 whole-repo descriptor。
func Build(ctx context.Context, opts BuildOptions) (*BuildResult, error) {
	cfg, configPath, err := project.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	if !cfg.IsProtoProject() {
		return nil, fmt.Errorf("descriptor build requires project.type=%s", project.ProjectTypeProto)
	}
	resolved, err := cfg.ResolveWithVersion(opts.Root, configPath, opts.Version)
	if err != nil {
		return nil, err
	}
	source := firstNonEmpty(opts.Source, cfg.Proto.Source, project.DefaultProtoSource)
	out := firstNonEmpty(opts.Out, resolved.DescriptorFile)
	out = absPath(opts.Root, out)
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return nil, err
	}
	buf := firstNonEmpty(opts.Buf, "buf")
	args := []string{
		"build",
		source,
		"--as-file-descriptor-set",
		"--exclude-source-info",
		"-o",
		out,
	}
	cmd := exec.CommandContext(ctx, buf, args...)
	cmd.Dir = opts.Root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("buf build failed: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	currentFile := resolved.CurrentDescriptorFile
	if currentFile != "" {
		if err = copyFile(out, currentFile); err != nil {
			return nil, err
		}
	}
	sha, size, err := fileDigest(out)
	if err != nil {
		return nil, err
	}
	return &BuildResult{
		ConfigPath:  configPath,
		Namespace:   resolved.Namespace,
		Version:     resolved.Version,
		Source:      source,
		File:        out,
		CurrentFile: currentFile,
		SHA256:      sha,
		Size:        size,
	}, nil
}

// Push 解析本地 descriptor 并按需上传到 S3 兼容对象存储。
func Push(ctx context.Context, opts PushOptions) (*PushResult, error) {
	cfg, configPath, err := project.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	if !cfg.IsProtoProject() {
		return nil, fmt.Errorf("descriptor push requires project.type=%s", project.ProjectTypeProto)
	}
	resolved, err := cfg.ResolveWithVersion(opts.Root, configPath, opts.Version)
	if err != nil {
		return nil, err
	}
	file := absPath(opts.Root, firstNonEmpty(opts.File, resolved.DescriptorFile))
	bucket := firstNonEmpty(opts.Bucket, env("FIREFLY_S3_BUCKET"), cfg.S3.Bucket)
	key := firstNonEmpty(opts.Key, resolved.ObjectKey)
	endpoint := firstNonEmpty(opts.Endpoint, env("FIREFLY_S3_ENDPOINT"), cfg.S3.Endpoint)
	region := firstNonEmpty(opts.Region, env("AWS_REGION"), env("AWS_DEFAULT_REGION"), cfg.S3.Region, project.DefaultS3Region)
	profile := firstNonEmpty(opts.Profile, cfg.S3.Profile)
	contentType := firstNonEmpty(opts.ContentType, cfg.Descriptor.ContentType, project.DefaultDescriptorContentType)
	forcePathStyle := opts.ForcePathStyle || envBool("FIREFLY_S3_FORCE_PATH_STYLE") || cfg.S3.ForcePathStyle
	versionedRef := firstNonEmpty(opts.VersionedRef, resolved.VersionedRef)
	currentKey := firstNonEmpty(opts.CurrentKey, resolved.CurrentObjectKey)
	currentFile := absPath(opts.Root, firstNonEmpty(opts.CurrentFile, resolved.CurrentDescriptorFile, file))
	currentRef := firstNonEmpty(opts.CurrentRef, resolved.CurrentRef)
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket is empty")
	}
	if key == "" {
		return nil, fmt.Errorf("s3 object key is empty")
	}
	if file == "" {
		return nil, fmt.Errorf("descriptor file is empty")
	}
	sha, size, err := fileDigest(file)
	if err != nil {
		return nil, err
	}
	result := &PushResult{
		ProjectType:   resolved.ProjectType,
		Namespace:     resolved.Namespace,
		Version:       resolved.Version,
		File:          file,
		CurrentFile:   currentFile,
		Bucket:        bucket,
		Key:           key,
		CurrentKey:    currentKey,
		VersionedRef:  versionedRef,
		CurrentRef:    currentRef,
		SHA256:        sha,
		Size:          size,
		DryRun:        opts.DryRun,
		PushedCurrent: cfg.IsProtoProject() && !opts.SkipCurrentObject,
	}
	if opts.DryRun {
		return result, nil
	}
	awsCfg, err := loadAWSConfig(ctx, awsConfigInput{
		Region:          region,
		Profile:         profile,
		AccessKeyID:     opts.AccessKeyID,
		SecretAccessKey: opts.SecretAccessKey,
		SessionToken:    opts.SessionToken,
	})
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})
	if err = putObject(ctx, client, bucket, key, file, contentType); err != nil {
		return nil, err
	}
	if cfg.IsProtoProject() && !opts.SkipCurrentObject {
		if currentKey == "" {
			return nil, fmt.Errorf("current s3 object key is empty")
		}
		uploadCurrentFile := currentFile
		if _, statErr := os.Stat(uploadCurrentFile); statErr != nil {
			uploadCurrentFile = file
		}
		if err = putObject(ctx, client, bucket, currentKey, uploadCurrentFile, contentType); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Publish 编排 proto descriptor build、push 和 Consul current 发布。
func Publish(ctx context.Context, opts PublishOptions) (*PublishResult, error) {
	cfg, configPath, err := project.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	if !cfg.IsProtoProject() {
		return nil, fmt.Errorf("descriptor publish requires project.type=%s", project.ProjectTypeProto)
	}
	var buildResult *BuildResult
	if !opts.SkipBuild {
		buildResult, err = Build(ctx, BuildOptions{
			Root:    opts.Root,
			Version: opts.Version,
			Source:  opts.Source,
			Out:     opts.Out,
			Buf:     opts.Buf,
		})
		if err != nil {
			return nil, err
		}
		if opts.File == "" {
			opts.File = buildResult.File
		}
	}
	pushResult, err := Push(ctx, opts.PushOptions)
	if err != nil {
		return nil, err
	}
	resolved, err := cfg.ResolveWithVersion(opts.Root, configPath, pushResult.Version)
	if err != nil {
		return nil, err
	}
	publishedAt := time.Now().UTC()
	sourceRevision := firstNonEmpty(opts.SourceRevision, readGitRevision(ctx, opts.Root))
	doc := descriptorCurrentDocument{
		Schema:         descriptorCurrentSchema,
		Namespace:      resolved.Namespace,
		Version:        pushResult.Version,
		Ref:            pushResult.VersionedRef,
		CurrentRef:     pushResult.CurrentRef,
		SHA256:         pushResult.SHA256,
		ProtoRepo:      cfg.Proto.Repo,
		SourceRevision: sourceRevision,
		PublishedAt:    publishedAt.Format(time.RFC3339),
	}
	currentJSON, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	consulAddress := firstNonEmpty(opts.ConsulAddress, cfg.Consul.Address)
	result := &PublishResult{
		Build:         buildResult,
		Push:          pushResult,
		CurrentKey:    resolved.DescriptorCurrentKey,
		ConsulAddress: consulAddress,
		CurrentJSON:   currentJSON,
		DryRun:        opts.DryRun,
		PublishedAt:   publishedAt,
	}
	if opts.SkipConsul || opts.DryRun {
		return result, nil
	}
	if consulAddress == "" {
		return nil, fmt.Errorf("consul.address is empty")
	}
	if resolved.DescriptorCurrentKey == "" {
		return nil, fmt.Errorf("descriptor current key is empty")
	}
	if err = putConsulKV(ctx, consulAddress, resolved.DescriptorCurrentKey, currentJSON); err != nil {
		return nil, err
	}
	return result, nil
}

// awsConfigInput 保存 AWS SDK 配置加载参数。
type awsConfigInput struct {
	// Region 是 S3 region。
	Region string
	// Profile 是 AWS shared config profile。
	Profile string
	// AccessKeyID 是显式 access key id。
	AccessKeyID string
	// SecretAccessKey 是显式 secret access key。
	SecretAccessKey string
	// SessionToken 是 STS 临时凭证 token。
	SessionToken string
}

func putObject(ctx context.Context, client *s3.Client, bucket, key, file, contentType string) error {
	body, err := os.Open(file)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

// loadAWSConfig 加载 AWS SDK 配置并保留默认凭证链。
func loadAWSConfig(ctx context.Context, input awsConfigInput) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(input.Region),
	}
	if input.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(input.Profile))
	}
	if input.AccessKeyID != "" || input.SecretAccessKey != "" || input.SessionToken != "" {
		if input.AccessKeyID == "" || input.SecretAccessKey == "" {
			return aws.Config{}, fmt.Errorf("access key id and secret access key must be set together")
		}
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			input.AccessKeyID,
			input.SecretAccessKey,
			input.SessionToken,
		)))
	}
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

func putConsulKV(ctx context.Context, address, key string, value []byte) error {
	endpoint := strings.TrimRight(address, "/") + "/v1/kv/" + strings.TrimLeft(key, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(value))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("consul kv put failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func readGitRevision(ctx context.Context, root string) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err = os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func absPath(root, path string) string {
	path = strings.TrimSpace(path)
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

// fileDigest 计算文件 sha256 和字节数。
func fileDigest(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("descriptor file not found: %s", path)
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

// env 读取环境变量并去掉首尾空白。
func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// envBool 解析常见布尔环境变量值。
func envBool(key string) bool {
	switch strings.ToLower(env(key)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
