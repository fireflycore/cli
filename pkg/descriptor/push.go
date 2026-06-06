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
	"github.com/fireflycore/cli/pkg/project"
)

// PushOptions 表示 descriptor push 命令传入的全部可覆盖参数。
type PushOptions struct {
	// Root 是业务服务仓库根目录，默认等于命令执行时的当前目录。
	Root string
	// File 用于显式指定本地 descriptor 文件路径。
	File string
	// Bucket 用于显式指定上传目标 bucket。
	Bucket string
	// Key 用于显式指定上传目标 object key。
	Key string
	// Endpoint 用于显式指定 S3 兼容服务地址。
	Endpoint string
	// Region 用于显式指定 S3 region。
	Region string
	// Profile 用于指定 AWS shared config/profile。
	Profile string
	// ForcePathStyle 控制是否使用 path-style 地址访问 S3。
	ForcePathStyle bool
	// ContentType 用于覆盖上传对象的 content type。
	ContentType string
	// DescriptorRef 用于覆盖命令最终打印的 descriptor_ref。
	DescriptorRef string
	// AccessKeyID 是显式传入的访问密钥 ID。
	AccessKeyID string
	// SecretAccessKey 是显式传入的访问密钥 secret。
	SecretAccessKey string
	// SessionToken 是显式传入的 STS 临时凭证 token。
	SessionToken string
	// DryRun 表示只解析和计算摘要，不执行真实上传。
	DryRun bool
}

// PushResult 表示 descriptor push 解析和上传后的结果。
type PushResult struct {
	// File 是最终使用的本地 descriptor 文件路径。
	File string
	// Bucket 是最终使用的上传目标 bucket。
	Bucket string
	// Key 是最终使用的上传目标 object key。
	Key string
	// DescriptorRef 是供 gateway 读取 descriptor 的最终引用地址。
	DescriptorRef string
	// SHA256 是本地 descriptor 文件的 sha256 摘要。
	SHA256 string
	// Size 是本地 descriptor 文件字节数。
	Size int64
	// DryRun 表示本次执行没有执行真实上传。
	DryRun bool
}

// Push 读取项目元信息，解析 descriptor 文件和 S3 目标，并按需上传对象。
func Push(ctx context.Context, opts PushOptions) (*PushResult, error) {
	// 读取当前业务仓库下的 .firefly/project.yaml。
	cfg, path, err := project.Load(opts.Root)
	if err != nil {
		return nil, err
	}
	// 根据 bootstrapConf.app.version 解析版本、文件路径、object key 和 descriptor_ref。
	resolved, err := cfg.Resolve(opts.Root, path)
	if err != nil {
		return nil, err
	}

	// 本地 descriptor 文件优先使用命令参数，其次使用项目配置推导值。
	file := firstNonEmpty(opts.File, resolved.DescriptorFile)
	// bucket 优先使用命令参数，其次使用环境变量，最后使用项目配置。
	bucket := firstNonEmpty(opts.Bucket, env("FIREFLY_S3_BUCKET"), cfg.S3.Bucket)
	// object key 优先使用命令参数，其次使用项目配置推导值。
	key := firstNonEmpty(opts.Key, resolved.ObjectKey)
	// endpoint 优先使用命令参数，其次使用环境变量，最后使用项目配置。
	endpoint := firstNonEmpty(opts.Endpoint, env("FIREFLY_S3_ENDPOINT"), cfg.S3.Endpoint)
	// region 优先使用命令参数，其次使用 AWS 环境变量，最后使用项目配置和默认值。
	region := firstNonEmpty(opts.Region, env("AWS_REGION"), env("AWS_DEFAULT_REGION"), cfg.S3.Region, project.DefaultS3Region)
	// profile 来自命令参数或项目配置。
	profile := firstNonEmpty(opts.Profile, cfg.S3.Profile)
	// content type 来自命令参数、项目配置或默认二进制类型。
	contentType := firstNonEmpty(opts.ContentType, cfg.Descriptor.ContentType, project.DefaultContentType)
	// path-style 可以由命令参数、环境变量或项目配置打开。
	forcePathStyle := opts.ForcePathStyle || envBool("FIREFLY_S3_FORCE_PATH_STYLE") || cfg.S3.ForcePathStyle
	// descriptor_ref 只用于输出，命令参数可以覆盖项目配置推导值。
	descriptorRef := firstNonEmpty(opts.DescriptorRef, resolved.DescriptorRef)

	// bucket 为空时无法构造 PutObject 请求。
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket is empty")
	}
	// object key 为空时无法确定上传目标路径。
	if key == "" {
		return nil, fmt.Errorf("s3 object key is empty")
	}
	// file 为空时说明配置没有成功推导本地 descriptor 路径。
	if file == "" {
		return nil, fmt.Errorf("descriptor file is empty")
	}

	// 先计算文件摘要和大小，dry-run 也需要输出这些信息。
	sha, size, err := fileDigest(file)
	if err != nil {
		return nil, err
	}

	// 汇总命令结果，后续无论 dry-run 还是真实上传都复用这份输出。
	result := &PushResult{
		File:          file,
		Bucket:        bucket,
		Key:           key,
		DescriptorRef: descriptorRef,
		SHA256:        sha,
		Size:          size,
		DryRun:        opts.DryRun,
	}
	// dry-run 到这里结束，不创建 S3 客户端，也不读取凭证。
	if opts.DryRun {
		return result, nil
	}

	// 打开 descriptor 文件作为 PutObject 请求体。
	body, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	// 上传结束后关闭文件句柄。
	defer body.Close()

	// 加载 AWS SDK 配置，支持 profile、默认凭证链和显式 STS 凭证。
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

	// 使用 AWS SDK 构造 S3 客户端；endpoint 允许接入 OSS、COS、MinIO 等兼容实现。
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// path-style 主要服务于 MinIO 或自建对象存储。
		o.UsePathStyle = forcePathStyle
		// endpoint 非空时覆盖 AWS 默认 endpoint。
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// 执行对象上传，descriptor push 不负责生成或修改 descriptor 内容。
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	// 上传成功后返回已经在本地计算好的结果。
	return result, nil
}

// awsConfigInput 是加载 AWS SDK 配置时需要的输入集合。
type awsConfigInput struct {
	// Region 是最终传给 AWS SDK 的 region。
	Region string
	// Profile 是 AWS shared config/profile 名称。
	Profile string
	// AccessKeyID 是显式静态凭证的 key id。
	AccessKeyID string
	// SecretAccessKey 是显式静态凭证的 secret。
	SecretAccessKey string
	// SessionToken 是 STS 临时凭证 token，可为空。
	SessionToken string
}

// loadAWSConfig 组装 AWS SDK 配置，保留 SDK 默认 credential chain。
func loadAWSConfig(ctx context.Context, input awsConfigInput) (aws.Config, error) {
	// region 总是写入加载参数，即使为空也交给 SDK 后续处理。
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(input.Region),
	}
	// profile 非空时从共享配置中读取对应 profile。
	if input.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(input.Profile))
	}
	// 任一显式凭证字段存在时，切换为静态凭证 provider。
	if input.AccessKeyID != "" || input.SecretAccessKey != "" || input.SessionToken != "" {
		// access key id 和 secret 必须成对出现，session token 可以单独为空。
		if input.AccessKeyID == "" || input.SecretAccessKey == "" {
			return aws.Config{}, fmt.Errorf("access key id and secret access key must be set together")
		}
		// 静态凭证 provider 同时兼容长期凭证和 STS 临时凭证。
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			input.AccessKeyID,
			input.SecretAccessKey,
			input.SessionToken,
		)))
	}
	// 其余凭证来源继续交给 AWS SDK 默认链处理。
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

// fileDigest 读取 descriptor 文件并返回 sha256 摘要与字节数。
func fileDigest(path string) (string, int64, error) {
	// 打开本地 descriptor 文件；不存在时给出明确路径。
	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("descriptor file not found: %s", path)
	}
	// 计算结束后关闭文件句柄。
	defer file.Close()

	// 创建 sha256 hash writer。
	hash := sha256.New()
	// 将文件内容流式写入 hash，并同时得到字节数。
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	// 将摘要转成十六进制字符串，便于命令行输出和人工核对。
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

// env 读取环境变量并去掉首尾空白。
func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// envBool 将常见真值字符串解析为布尔值。
func envBool(key string) bool {
	// 兼容 shell 中常见的布尔写法。
	switch strings.ToLower(env(key)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// firstNonEmpty 返回第一个去掉空白后仍非空的字符串。
func firstNonEmpty(values ...string) string {
	// 按调用方传入顺序体现配置优先级。
	for _, value := range values {
		// 每个候选值都先 trim，避免空白字符串误判为有效值。
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	// 所有候选值都为空时返回空字符串。
	return ""
}
