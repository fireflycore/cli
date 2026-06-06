package descriptor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fireflycore/cli/internal/project"
)

// PushOptions 表示 descriptor push 命令传入的参数。
type PushOptions struct {
	// Root 是业务服务仓库根目录。
	Root string
	// File 显式覆盖本地 descriptor 文件路径。
	File string
	// Bucket 显式覆盖目标 bucket。
	Bucket string
	// Key 显式覆盖目标 object key。
	Key string
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
	// DescriptorRef 覆盖命令最终打印的 descriptor_ref。
	DescriptorRef string
	// AccessKeyID 是显式传入的 access key id。
	AccessKeyID string
	// SecretAccessKey 是显式传入的 secret access key。
	SecretAccessKey string
	// SessionToken 是 STS 临时凭证 token。
	SessionToken string
	// DryRun 表示只解析和计算摘要，不上传对象。
	DryRun bool
}

// PushResult 表示 descriptor push 的解析和上传结果。
type PushResult struct {
	// File 是最终使用的本地 descriptor 文件。
	File string
	// Bucket 是最终使用的目标 bucket。
	Bucket string
	// Key 是最终使用的 object key。
	Key string
	// DescriptorRef 是最终输出的 descriptor_ref。
	DescriptorRef string
	// SHA256 是 descriptor 文件 sha256 摘要。
	SHA256 string
	// Size 是 descriptor 文件字节数。
	Size int64
	// DryRun 表示本次没有真实上传。
	DryRun bool
}

// Push 解析本地 descriptor 并按需上传到 S3 兼容对象存储。
func Push(ctx context.Context, opts PushOptions) (*PushResult, error) {
	// 读取项目配置。
	cfg, configPath, err := project.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	// 按 bootstrapConf.app.version 解析版本和 descriptor 路径。
	resolved, err := cfg.Resolve(opts.Root, configPath)
	if err != nil {
		return nil, err
	}
	// 按命令参数、环境变量、项目配置的顺序合并上传参数。
	file := firstNonEmpty(opts.File, resolved.DescriptorFile)
	bucket := firstNonEmpty(opts.Bucket, env("FIREFLY_S3_BUCKET"), cfg.S3.Bucket)
	key := firstNonEmpty(opts.Key, resolved.ObjectKey)
	endpoint := firstNonEmpty(opts.Endpoint, env("FIREFLY_S3_ENDPOINT"), cfg.S3.Endpoint)
	region := firstNonEmpty(opts.Region, env("AWS_REGION"), env("AWS_DEFAULT_REGION"), cfg.S3.Region, project.DefaultS3Region)
	profile := firstNonEmpty(opts.Profile, cfg.S3.Profile)
	contentType := firstNonEmpty(opts.ContentType, cfg.Descriptor.ContentType, project.DefaultDescriptorContentType)
	forcePathStyle := opts.ForcePathStyle || envBool("FIREFLY_S3_FORCE_PATH_STYLE") || cfg.S3.ForcePathStyle
	descriptorRef := firstNonEmpty(opts.DescriptorRef, resolved.DescriptorRef)
	// 上传目标必须完整。
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket is empty")
	}
	if key == "" {
		return nil, fmt.Errorf("s3 object key is empty")
	}
	if file == "" {
		return nil, fmt.Errorf("descriptor file is empty")
	}
	// dry-run 也需要先确认文件存在并计算摘要。
	sha, size, err := fileDigest(file)
	if err != nil {
		return nil, err
	}
	result := &PushResult{
		File:          file,
		Bucket:        bucket,
		Key:           key,
		DescriptorRef: descriptorRef,
		SHA256:        sha,
		Size:          size,
		DryRun:        opts.DryRun,
	}
	// dry-run 不读取凭证、不创建 S3 client。
	if opts.DryRun {
		return result, nil
	}
	// 打开文件作为 PutObject body。
	body, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	// 加载 AWS SDK 配置。
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
	// 创建 S3 client，endpoint 非空时接入兼容对象存储。
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})
	// 执行上传。
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
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

// loadAWSConfig 加载 AWS SDK 配置并保留默认凭证链。
func loadAWSConfig(ctx context.Context, input awsConfigInput) (aws.Config, error) {
	// region 总是传入 SDK。
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(input.Region),
	}
	// profile 非空时使用共享配置。
	if input.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(input.Profile))
	}
	// 显式凭证存在时使用静态 provider。
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

// fileDigest 计算文件 sha256 和字节数。
func fileDigest(path string) (string, int64, error) {
	// 打开 descriptor 文件。
	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("descriptor file not found: %s", path)
	}
	defer file.Close()
	// 流式计算摘要。
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
