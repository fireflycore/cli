package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fireflycore/cli/internal/descriptor"
	"github.com/fireflycore/cli/internal/project"
	"github.com/fireflycore/cli/internal/template"
	"github.com/spf13/cobra"
)

const (
	// AppName 是 CLI 在缓存目录中的名字。
	AppName = "firefly"
	// Version 是当前 CLI 版本。
	Version = "v0.0.9"
)

// Runtime 保存命令执行时需要的外部上下文。
type Runtime struct {
	// WorkDir 是用户执行命令时所在目录。
	WorkDir string
	// CacheDir 是 CLI 缓存目录。
	CacheDir string
	// In 是命令输入流。
	In io.Reader
	// Out 是命令标准输出流。
	Out io.Writer
	// Err 是命令错误输出流。
	Err io.Writer
}

// NewRuntime 从当前进程环境构造运行时上下文。
func NewRuntime() (Runtime, error) {
	// 获取当前工作目录。
	workDir, err := os.Getwd()
	if err != nil {
		return Runtime{}, err
	}
	// 获取用户缓存目录。
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Runtime{}, err
	}
	return Runtime{
		WorkDir:  workDir,
		CacheDir: filepath.Join(cacheRoot, AppName),
		In:       os.Stdin,
		Out:      os.Stdout,
		Err:      os.Stderr,
	}, nil
}

// NewCommand 创建完整 Cobra 命令树。
func NewCommand(runtime Runtime) *cobra.Command {
	// 根命令只负责声明 CLI 元信息和挂载子命令。
	root := &cobra.Command{
		Use:     "firefly",
		Short:   "Firefly Go microservice engineering helper.",
		Version: Version,
	}
	// 显式设置输入输出，便于测试和组合。
	root.SetIn(runtime.In)
	root.SetOut(runtime.Out)
	root.SetErr(runtime.Err)
	// 挂载当前规划保留的三个能力。
	root.AddCommand(newCreateCommand(runtime))
	root.AddCommand(newProjectCommand(runtime))
	root.AddCommand(newDescriptorCommand(runtime))
	return root
}

// Execute 构造并执行命令树。
func Execute(ctx context.Context, runtime Runtime, args []string) error {
	// 创建命令树。
	cmd := NewCommand(runtime)
	// 调用方可以显式传入 args，main 传 nil 时使用 os.Args。
	if args != nil {
		cmd.SetArgs(args)
	}
	return cmd.ExecuteContext(ctx)
}

// createFlags 保存 create 命令参数。
type createFlags struct {
	// name 是项目目录名。
	name string
	// language 是模板语言。
	language string
	// templateVersion 是模板版本。
	templateVersion string
	// module 是 Go module 名。
	module string
	// appID 是 Firefly app id。
	appID string
	// service 是服务名。
	service string
	// nonInteractive 控制是否禁止交互式询问。
	nonInteractive bool
}

// newCreateCommand 创建 create 命令。
func newCreateCommand(runtime Runtime) *cobra.Command {
	// flags 保存本命令生命周期内的参数。
	flags := createFlags{
		language:        template.GoLanguage,
		templateVersion: template.LatestVersion,
	}
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a Firefly service project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// positional name 作为 --name 的轻量替代。
			if flags.name == "" && len(args) > 0 {
				flags.name = args[0]
			}
			// 需要交互时从 stdin 读取缺失字段。
			if flags.name == "" && !flags.nonInteractive {
				name, err := prompt(cmd.InOrStdin(), cmd.OutOrStdout(), "Project name")
				if err != nil {
					return err
				}
				flags.name = name
			}
			// 禁止交互时必须显式给出项目名。
			if flags.name == "" {
				return fmt.Errorf("project name is required")
			}
			// 调用模板创建服务项目。
			result, err := template.Create(cmd.Context(), template.CreateOptions{
				Root:            runtime.WorkDir,
				CacheDir:        runtime.CacheDir,
				Name:            flags.name,
				Language:        flags.language,
				TemplateVersion: flags.templateVersion,
				Module:          flags.module,
				AppID:           flags.appID,
				Service:         flags.service,
			})
			if err != nil {
				return err
			}
			// 输出机器可读的创建结果。
			fmt.Fprintf(cmd.OutOrStdout(), "created: %s\n", result.Path)
			fmt.Fprintf(cmd.OutOrStdout(), "template: %s\n", result.TemplateVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "module: %s\n", result.Module)
			return nil
		},
	}
	// 声明 create 参数。
	cmd.Flags().StringVar(&flags.name, "name", "", "project directory name")
	cmd.Flags().StringVar(&flags.language, "language", template.GoLanguage, "development language")
	cmd.Flags().StringVar(&flags.templateVersion, "template-version", template.LatestVersion, "template version, defaults to latest")
	cmd.Flags().StringVar(&flags.module, "module", "", "Go module name")
	cmd.Flags().StringVar(&flags.appID, "app-id", "", "Firefly app id")
	cmd.Flags().StringVar(&flags.service, "service", "", "service name")
	cmd.Flags().BoolVar(&flags.nonInteractive, "non-interactive", false, "disable interactive prompts")
	return cmd
}

// newProjectCommand 创建 project 命令组。
func newProjectCommand(runtime Runtime) *cobra.Command {
	// project 是本地项目元信息命令组。
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Firefly project metadata",
	}
	cmd.AddCommand(newProjectInitCommand(runtime))
	cmd.AddCommand(newProjectInfoCommand(runtime))
	cmd.AddCommand(newProjectCheckCommand(runtime))
	return cmd
}

// newProjectInitCommand 创建 project init 命令。
func newProjectInitCommand(runtime Runtime) *cobra.Command {
	// opts 直接复用 project 包的初始化参数。
	opts := project.InitOptions{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create .firefly/project.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 项目配置始终写入当前工作目录。
			opts.Root = runtime.WorkDir
			cfg, err := project.NewConfig(opts)
			if err != nil {
				return err
			}
			path, err := project.Save(runtime.WorkDir, cfg, opts.Overwrite)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created: %s\n", path)
			return nil
		},
	}
	// 声明 project init 参数。
	cmd.Flags().StringVar(&opts.ServiceName, "service", "", "service name, defaults to current directory name")
	cmd.Flags().StringVar(&opts.AppID, "app-id", "", "Firefly app id, defaults to service name")
	cmd.Flags().StringVar(&opts.Namespace, "namespace", project.DefaultNamespace, "service namespace")
	cmd.Flags().StringVar(&opts.Language, "language", project.DefaultLanguage, "project language")
	cmd.Flags().StringVar(&opts.Module, "module", "", "Go module name, defaults to go.mod")
	cmd.Flags().StringVar(&opts.BootstrapFile, "bootstrap-file", project.DefaultBootstrapFile, "bootstrap config file path")
	cmd.Flags().StringVar(&opts.VersionPath, "version-path", project.DefaultBootstrapVersionPath, "version field path in bootstrap config")
	cmd.Flags().StringVar(&opts.DescriptorDir, "descriptor-dir", project.DefaultDescriptorDir, "descriptor output directory")
	cmd.Flags().StringVar(&opts.FileTemplate, "file-template", project.DefaultDescriptorFileTemplate, "descriptor file name template")
	cmd.Flags().StringVar(&opts.ObjectKeyTemplate, "object-key-template", project.DefaultObjectKeyTemplate, "S3 object key template")
	cmd.Flags().StringVar(&opts.RefTemplate, "descriptor-ref-template", "", "descriptor_ref template")
	cmd.Flags().StringVar(&opts.DescriptorRef, "descriptor-ref", "", "fixed descriptor_ref")
	cmd.Flags().StringVar(&opts.ContentType, "content-type", project.DefaultDescriptorContentType, "descriptor content type")
	cmd.Flags().StringVar(&opts.S3Profile, "s3-profile", "", "AWS shared config profile")
	cmd.Flags().StringVar(&opts.S3Region, "s3-region", project.DefaultS3Region, "S3 region")
	cmd.Flags().StringVar(&opts.S3Endpoint, "s3-endpoint", "", "S3-compatible endpoint")
	cmd.Flags().StringVar(&opts.S3Bucket, "s3-bucket", project.DefaultS3Bucket, "S3 bucket")
	cmd.Flags().BoolVar(&opts.ForcePathStyle, "s3-force-path-style", false, "use S3 path-style addressing")
	cmd.Flags().BoolVar(&opts.Overwrite, "overwrite", false, "overwrite existing project config")
	return cmd
}

// newProjectInfoCommand 创建 project info 命令。
func newProjectInfoCommand(runtime Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show Firefly project metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 读取项目配置。
			cfg, path, err := project.Load(runtime.WorkDir)
			if err != nil {
				return err
			}
			// 尝试解析版本和 descriptor 路径。
			resolved, resolveErr := cfg.Resolve(runtime.WorkDir, path)
			// 输出基础元信息。
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "config: %s\n", path)
			fmt.Fprintf(out, "service: %s\n", cfg.Service.Name)
			fmt.Fprintf(out, "app_id: %s\n", cfg.Service.AppID)
			fmt.Fprintf(out, "namespace: %s\n", cfg.Service.Namespace)
			fmt.Fprintf(out, "language: %s\n", cfg.Service.Language)
			fmt.Fprintf(out, "module: %s\n", cfg.Service.Module)
			if resolveErr != nil {
				fmt.Fprintf(out, "version_error: %s\n", resolveErr)
			} else {
				fmt.Fprintf(out, "version: %s\n", resolved.Version)
				fmt.Fprintf(out, "descriptor.file: %s\n", resolved.DescriptorFile)
				fmt.Fprintf(out, "descriptor.object_key: %s\n", resolved.ObjectKey)
				fmt.Fprintf(out, "descriptor_ref: %s\n", resolved.DescriptorRef)
			}
			// 输出 Firefly 依赖版本。
			writeDependencies(out, project.ReadFireflyModules(runtime.WorkDir))
			return nil
		},
	}
}

// newProjectCheckCommand 创建 project check 命令。
func newProjectCheckCommand(runtime Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Run local Firefly project checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 执行本地检查，不连接运行时系统。
			results, _, _ := project.Check(runtime.WorkDir)
			hasFailed := false
			for _, result := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", result.Status, result.Name, result.Message)
				if result.Status == project.CheckFailed {
					hasFailed = true
				}
			}
			if hasFailed {
				return fmt.Errorf("project check failed")
			}
			return nil
		},
	}
}

// newDescriptorCommand 创建 descriptor 命令组。
func newDescriptorCommand(runtime Runtime) *cobra.Command {
	// descriptor 当前只保留 push。
	cmd := &cobra.Command{
		Use:   "descriptor",
		Short: "Publish Firefly gateway descriptors",
	}
	cmd.AddCommand(newDescriptorPushCommand(runtime))
	return cmd
}

// newDescriptorPushCommand 创建 descriptor push 命令。
func newDescriptorPushCommand(runtime Runtime) *cobra.Command {
	// opts 保存本次 push 的参数。
	opts := descriptor.PushOptions{}
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push an existing descriptor file to S3-compatible storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			// descriptor push 必须在业务服务仓库内执行。
			opts.Root = runtime.WorkDir
			result, err := descriptor.Push(cmd.Context(), opts)
			if err != nil {
				return err
			}
			// 输出上传结果。
			out := cmd.OutOrStdout()
			if result.DryRun {
				fmt.Fprintln(out, "dry_run: true")
			}
			fmt.Fprintf(out, "file: %s\n", result.File)
			fmt.Fprintf(out, "size: %d\n", result.Size)
			fmt.Fprintf(out, "sha256: %s\n", result.SHA256)
			fmt.Fprintf(out, "bucket: %s\n", result.Bucket)
			fmt.Fprintf(out, "key: %s\n", result.Key)
			if result.DescriptorRef != "" {
				fmt.Fprintf(out, "descriptor_ref: %s\n", result.DescriptorRef)
			}
			if !result.DryRun {
				fmt.Fprintln(out, "pushed: true")
			}
			return nil
		},
	}
	// 声明 descriptor push 参数。
	cmd.Flags().StringVar(&opts.File, "file", "", "descriptor file path, defaults to project config and service version")
	cmd.Flags().StringVar(&opts.Bucket, "bucket", "", "S3 bucket, defaults to project config or FIREFLY_S3_BUCKET")
	cmd.Flags().StringVar(&opts.Key, "key", "", "S3 object key, defaults to {service}/{version}.pb")
	cmd.Flags().StringVar(&opts.Endpoint, "endpoint", "", "S3-compatible endpoint, defaults to project config or FIREFLY_S3_ENDPOINT")
	cmd.Flags().StringVar(&opts.Region, "region", "", "S3 region, defaults to AWS_REGION or project config")
	cmd.Flags().StringVar(&opts.Profile, "profile", "", "AWS shared config profile")
	cmd.Flags().BoolVar(&opts.ForcePathStyle, "force-path-style", false, "use S3 path-style addressing")
	cmd.Flags().StringVar(&opts.ContentType, "content-type", "", "descriptor content type")
	cmd.Flags().StringVar(&opts.DescriptorRef, "descriptor-ref", "", "descriptor_ref URL to print")
	cmd.Flags().StringVar(&opts.AccessKeyID, "access-key-id", "", "S3 access key id")
	cmd.Flags().StringVar(&opts.SecretAccessKey, "secret-access-key", "", "S3 secret access key")
	cmd.Flags().StringVar(&opts.SessionToken, "session-token", "", "STS temporary credential session token")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "resolve and hash the descriptor without uploading")
	return cmd
}

// prompt 读取一行交互式输入。
func prompt(in io.Reader, out io.Writer, label string) (string, error) {
	// 打印 prompt。
	fmt.Fprintf(out, "%s: ", label)
	// 读取用户输入。
	reader := bufio.NewReader(in)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

// writeDependencies 按名称排序输出 Firefly 依赖版本。
func writeDependencies(out io.Writer, deps map[string]string) {
	// 没有依赖时不输出 dependencies 区块。
	if len(deps) == 0 {
		return
	}
	// 稳定排序便于脚本和人工比较。
	keys := make([]string, 0, len(deps))
	for key := range deps {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Fprintln(out, "dependencies:")
	for _, key := range keys {
		fmt.Fprintf(out, "  %s: %s\n", strings.TrimPrefix(key, "github.com/fireflycore/"), deps[key])
	}
}
