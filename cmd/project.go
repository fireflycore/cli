package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fireflycore/cli/pkg/project"
	"github.com/spf13/cobra"
)

// projectInitOpts 保存 project init 命令收集到的初始化参数。
var projectInitOpts project.InitOptions

// projectCmd 是项目元信息命令组，只管理本地 .firefly/project.yaml。
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "管理 Firefly 项目元信息",
}

// projectInitCmd 负责在当前仓库创建 .firefly/project.yaml。
var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "创建 .firefly/project.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取当前工作目录，项目配置始终写入当前仓库。
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// 将根目录写回初始化参数，便于 project 包推导默认服务名和 go.mod。
		projectInitOpts.Root = root

		// 根据命令参数和默认值构造项目配置。
		cfg, err := project.NewConfig(projectInitOpts)
		if err != nil {
			return err
		}

		// 将配置写入 .firefly/project.yaml。
		path, err := project.Save(root, cfg, projectInitOpts.Overwrite)
		if err != nil {
			return err
		}

		// 输出创建结果，方便脚本模式读取。
		fmt.Printf("created %s\n", path)
		return nil
	},
}

// projectInfoCmd 负责展示当前项目配置、服务版本和 Firefly 依赖版本。
var projectInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "展示 Firefly 项目元信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取当前仓库目录。
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// 读取 .firefly/project.yaml。
		cfg, path, err := project.Load(root)
		if err != nil {
			return err
		}

		// 解析 bootstrapConf.app.version 和 descriptor 路径。
		resolved, resolveErr := cfg.Resolve(root, path)
		// 从 go.mod 中提取 fireflycore 依赖，便于人工确认基线版本。
		deps := project.ReadFireflyModules(root)

		// 输出项目配置文件路径。
		fmt.Printf("config: %s\n", path)
		// 输出服务名。
		fmt.Printf("service: %s\n", cfg.Service.Name)
		// 输出 Firefly app id。
		fmt.Printf("app_id: %s\n", cfg.Service.AppID)
		// 输出服务命名空间。
		fmt.Printf("namespace: %s\n", cfg.Service.Namespace)
		// 输出项目语言。
		fmt.Printf("language: %s\n", cfg.Service.Language)
		// 输出 Go module 名。
		fmt.Printf("module: %s\n", cfg.Service.Module)

		// 如果版本解析失败，展示错误但不阻断基础信息输出。
		if resolveErr != nil {
			fmt.Printf("version_error: %s\n", resolveErr)
		} else {
			// 输出 bootstrap 中读取到的服务版本。
			fmt.Printf("version: %s\n", resolved.Version)
			// 输出本地 descriptor 文件路径。
			fmt.Printf("descriptor.file: %s\n", resolved.DescriptorFile)
			// 输出 S3 对象 key。
			fmt.Printf("descriptor.object_key: %s\n", resolved.ObjectKey)
			// 输出 api-gateway 可访问的 descriptor_ref。
			fmt.Printf("descriptor_ref: %s\n", resolved.DescriptorRef)
		}

		// 如果 go.mod 中存在 fireflycore 依赖，则按名称排序后输出。
		if len(deps) > 0 {
			fmt.Println("dependencies:")
			keys := make([]string, 0, len(deps))
			for key := range deps {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("  %s: %s\n", strings.TrimPrefix(key, "github.com/fireflycore/"), deps[key])
			}
		}

		return nil
	},
}

// projectCheckCmd 负责执行不依赖运行时组件的本地静态检查。
var projectCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "执行 Firefly 项目本地检查",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取当前仓库目录。
		root, err := os.Getwd()
		if err != nil {
			return err
		}

		// 执行本地检查，检查逻辑不连接 sidecar、gateway 或 authz。
		results, _, _ := project.Check(root)
		// 记录是否出现失败项，warn 只提示不让命令失败。
		hasFailed := false
		for _, result := range results {
			fmt.Printf("[%s] %s: %s\n", result.Status, result.Name, result.Message)
			if result.Status == project.CheckFailed {
				hasFailed = true
			}
		}

		// 只有 failed 项会导致命令返回错误。
		if hasFailed {
			return fmt.Errorf("project check failed")
		}
		return nil
	},
}

func init() {
	// 将 project 命令组挂载到根命令。
	rootCmd.AddCommand(projectCmd)
	// 注册 project init 子命令。
	projectCmd.AddCommand(projectInitCmd)
	// 注册 project info 子命令。
	projectCmd.AddCommand(projectInfoCmd)
	// 注册 project check 子命令。
	projectCmd.AddCommand(projectCheckCmd)

	// --service 指定服务名，默认使用当前目录名。
	projectInitCmd.Flags().StringVar(&projectInitOpts.ServiceName, "service", "", "服务名，默认使用当前目录名")
	// --app-id 指定 Firefly 应用 ID，默认等于服务名。
	projectInitCmd.Flags().StringVar(&projectInitOpts.AppID, "app-id", "", "Firefly 应用 ID，默认等于服务名")
	// --namespace 指定服务命名空间。
	projectInitCmd.Flags().StringVar(&projectInitOpts.Namespace, "namespace", project.DefaultNamespace, "服务命名空间")
	// --language 指定项目语言。
	projectInitCmd.Flags().StringVar(&projectInitOpts.Language, "language", project.DefaultLanguage, "项目语言")
	// --module 指定 Go module 名，默认读取 go.mod。
	projectInitCmd.Flags().StringVar(&projectInitOpts.Module, "module", "", "Go module 名，默认读取 go.mod")
	// --bootstrap-file 指定启动配置文件路径。
	projectInitCmd.Flags().StringVar(&projectInitOpts.BootstrapFile, "bootstrap-file", project.DefaultBootstrapFile, "启动配置文件路径")
	// --version-path 指定版本字段路径，默认 app.version。
	projectInitCmd.Flags().StringVar(&projectInitOpts.VersionPath, "version-path", project.DefaultBootstrapVersion, "启动配置中的版本字段路径")
	// --descriptor-dir 指定 descriptor 本地目录。
	projectInitCmd.Flags().StringVar(&projectInitOpts.DescriptorDir, "descriptor-dir", project.DefaultDescriptorDir, "descriptor 输出目录")
	// --file-template 指定 descriptor 文件名模板。
	projectInitCmd.Flags().StringVar(&projectInitOpts.FileTemplate, "file-template", project.DefaultFileTemplate, "descriptor 文件名模板")
	// --object-key-template 指定 S3 对象 key 模板。
	projectInitCmd.Flags().StringVar(&projectInitOpts.ObjectKeyTemplate, "object-key-template", project.DefaultObjectKeyTemplate, "S3 对象 key 模板")
	// --descriptor-ref-template 指定 descriptor_ref 模板。
	projectInitCmd.Flags().StringVar(&projectInitOpts.RefTemplate, "descriptor-ref-template", "", "descriptor_ref 模板")
	// --descriptor-ref 指定固定 descriptor_ref。
	projectInitCmd.Flags().StringVar(&projectInitOpts.DescriptorRef, "descriptor-ref", "", "固定 descriptor_ref")
	// --content-type 指定上传对象的 content type。
	projectInitCmd.Flags().StringVar(&projectInitOpts.ContentType, "content-type", project.DefaultContentType, "descriptor 内容类型")
	// --s3-profile 指定 AWS 共享配置 profile。
	projectInitCmd.Flags().StringVar(&projectInitOpts.S3Profile, "s3-profile", "", "AWS 共享配置 profile")
	// --s3-region 指定 S3 region。
	projectInitCmd.Flags().StringVar(&projectInitOpts.S3Region, "s3-region", project.DefaultS3Region, "S3 region")
	// --s3-endpoint 指定 S3 兼容 endpoint。
	projectInitCmd.Flags().StringVar(&projectInitOpts.S3Endpoint, "s3-endpoint", "", "S3 兼容 endpoint")
	// --s3-bucket 指定 S3 bucket，默认 descriptor。
	projectInitCmd.Flags().StringVar(&projectInitOpts.S3Bucket, "s3-bucket", project.DefaultS3Bucket, "S3 bucket")
	// --s3-force-path-style 指定是否使用路径风格访问，MinIO 常用。
	projectInitCmd.Flags().BoolVar(&projectInitOpts.ForcePathStyle, "s3-force-path-style", false, "使用 S3 path-style 地址")
	// --overwrite 允许覆盖已有 .firefly/project.yaml。
	projectInitCmd.Flags().BoolVar(&projectInitOpts.Overwrite, "overwrite", false, "覆盖已有项目配置")
}
