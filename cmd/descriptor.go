package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fireflycore/cli/pkg/descriptor"
	"github.com/spf13/cobra"
)

// descriptorPushOpts 保存 descriptor push 命令收集到的上传参数。
var descriptorPushOpts descriptor.PushOptions

// descriptorCmd 是 descriptor 命令组，当前只保留 push 能力。
var descriptorCmd = &cobra.Command{
	Use:   "descriptor",
	Short: "Publish Firefly gateway descriptors",
}

// descriptorPushCmd 负责把已经由 Makefile 生成的 descriptor 上传到 S3 兼容存储。
var descriptorPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push an existing descriptor file to S3-compatible storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取当前仓库目录，descriptor push 必须在业务服务仓库内执行。
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// 将当前目录写入 options，供 descriptor 包读取 .firefly/project.yaml。
		descriptorPushOpts.Root = root

		// 执行 descriptor 解析、sha256 计算以及可选上传。
		result, err := descriptor.Push(context.Background(), descriptorPushOpts)
		if err != nil {
			return err
		}

		// dry-run 模式只解析和计算摘要，不执行 PutObject。
		if result.DryRun {
			fmt.Println("dry_run: true")
		}
		// 输出本地 descriptor 文件路径。
		fmt.Printf("file: %s\n", result.File)
		// 输出文件大小，便于发布记录留痕。
		fmt.Printf("size: %d\n", result.Size)
		// 输出 sha256，便于人工核对对象内容。
		fmt.Printf("sha256: %s\n", result.SHA256)
		// 输出目标 bucket。
		fmt.Printf("bucket: %s\n", result.Bucket)
		// 输出目标 object key。
		fmt.Printf("key: %s\n", result.Key)
		// 如果能推导 descriptor_ref，则输出给发布流程使用。
		if result.DescriptorRef != "" {
			fmt.Printf("descriptor_ref: %s\n", result.DescriptorRef)
		}
		// 非 dry-run 成功时明确输出 pushed 标记。
		if !result.DryRun {
			fmt.Println("pushed: true")
		}
		return nil
	},
}

func init() {
	// 将 descriptor 命令组挂载到根命令。
	rootCmd.AddCommand(descriptorCmd)
	// 注册 descriptor push 子命令。
	descriptorCmd.AddCommand(descriptorPushCmd)

	// --file 覆盖默认 descriptor 文件路径。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.File, "file", "", "descriptor file path, defaults to project config and service version")
	// --bucket 覆盖项目配置或 FIREFLY_S3_BUCKET 中的 bucket。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.Bucket, "bucket", "", "S3 bucket, defaults to project config or FIREFLY_S3_BUCKET")
	// --key 覆盖默认对象 key。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.Key, "key", "", "S3 object key, defaults to {service}/{version}.pb")
	// --endpoint 覆盖项目配置或 FIREFLY_S3_ENDPOINT 中的 endpoint。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.Endpoint, "endpoint", "", "S3-compatible endpoint, defaults to project config or FIREFLY_S3_ENDPOINT")
	// --region 覆盖 AWS_REGION 或项目配置中的 region。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.Region, "region", "", "S3 region, defaults to AWS_REGION or project config")
	// --profile 指定 AWS 共享配置 profile。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.Profile, "profile", "", "AWS shared config profile")
	// --force-path-style 强制使用路径风格地址，MinIO 常用。
	descriptorPushCmd.Flags().BoolVar(&descriptorPushOpts.ForcePathStyle, "force-path-style", false, "use S3 path-style addressing")
	// --content-type 覆盖上传对象的 content type。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.ContentType, "content-type", "", "descriptor content type")
	// --descriptor-ref 覆盖最终打印的 descriptor_ref。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.DescriptorRef, "descriptor-ref", "", "descriptor_ref URL to print")
	// --access-key-id 显式指定访问密钥 ID。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.AccessKeyID, "access-key-id", "", "S3 access key id")
	// --secret-access-key 显式指定访问密钥 secret。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.SecretAccessKey, "secret-access-key", "", "S3 secret access key")
	// --session-token 显式指定 STS 临时凭证 token。
	descriptorPushCmd.Flags().StringVar(&descriptorPushOpts.SessionToken, "session-token", "", "STS temporary credential session token")
	// --dry-run 只解析路径和计算摘要，不上传对象。
	descriptorPushCmd.Flags().BoolVar(&descriptorPushOpts.DryRun, "dry-run", false, "resolve and hash the descriptor without uploading")
}
