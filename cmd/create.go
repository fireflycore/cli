package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fireflycore/cli/pkg/repo"
	"github.com/fireflycore/cli/pkg/store"
	"github.com/fireflycore/cli/pkg/view"
	"github.com/spf13/cobra"
)

var (
	// createProjectName 保存非交互模式下传入的项目目录名。
	createProjectName string
	// createLanguage 保存目标开发语言，目前主线只支持 go。
	createLanguage string
	// createModule 保存生成项目的 Go module 名。
	createModule string
	// createAppID 保存写入 bootstrap.json 的 Firefly app.id。
	createAppID string
	// createServiceName 保存写入 bootstrap.json 的 service.name。
	createServiceName string
	// createNonInteractive 控制是否跳过 Bubble Tea 交互表单。
	createNonInteractive bool
)

// createNameReg 用于清理项目名中的非法字符，避免生成异常目录名。
var createNameReg = regexp.MustCompile("[^a-zA-Z0-9_-]+")

// createCmd 负责从 go-layout 模板创建新的业务服务项目。
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Firefly service project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 根据 flag 判断走交互式表单还是非交互式参数。
		cfg, err := resolveCreateConfig()
		if err != nil {
			return err
		}

		// 读取当前语言对应的缓存模板版本。
		v := store.Use.Config.Global.Version[cfg.Language]
		// 如果用户没有显式指定模板版本，则优先复用全局缓存版本。
		if v != "latest" && templateVersion == "latest" {
			templateVersion = v
		}

		// 构造模板仓库客户端，后续负责拉取模板和替换工程内容。
		rc, err := repo.New(&repo.ConfigEntity{
			Project:  cfg.Project,
			Language: cfg.Language,
			Version:  templateVersion,
			Module:   createModule,
			AppID:    createAppID,
			Service:  createServiceName,
		})
		if err != nil {
			return err
		}

		// 获取模板并初始化项目目录。
		if err = rc.GetRepo(); err != nil {
			return err
		}

		// 输出最终创建的项目名，方便脚本模式读取。
		fmt.Printf("created %s\n", cfg.Project)
		return nil
	},
}

func init() {
	// 将 create 命令挂载到根命令。
	rootCmd.AddCommand(createCmd)

	// --version 保留为历史兼容参数，语义等同于 --template-version。
	createCmd.Flags().StringVar(&templateVersion, "version", "latest", "template version, defaults to latest")
	// --template-version 是新文档推荐使用的模板版本参数。
	createCmd.Flags().StringVar(&templateVersion, "template-version", "latest", "template version, defaults to latest")
	// --name 指定生成的项目目录名。
	createCmd.Flags().StringVar(&createProjectName, "name", "", "project directory name")
	// --language 指定开发语言，目前主线仅支持 go。
	createCmd.Flags().StringVar(&createLanguage, "language", "go", "development language")
	// --module 指定 Go module 名，生成后会写入 go.mod 和 import path。
	createCmd.Flags().StringVar(&createModule, "module", "", "Go module name")
	// --app-id 指定 bootstrap.json 中的 app.id。
	createCmd.Flags().StringVar(&createAppID, "app-id", "", "Firefly app id")
	// --service 指定 bootstrap.json 中的 service.name。
	createCmd.Flags().StringVar(&createServiceName, "service", "", "service name")
	// --non-interactive 禁用交互式表单，适合脚本和 CI。
	createCmd.Flags().BoolVar(&createNonInteractive, "non-interactive", false, "disable interactive prompts")
}

// resolveCreateConfig 根据当前 flag 决定返回交互式表单结果或非交互式配置。
func resolveCreateConfig() (*view.CreateFormEntity, error) {
	// 只要显式启用非交互模式或传入项目名，就不再启动 TUI。
	if createNonInteractive || strings.TrimSpace(createProjectName) != "" {
		// 清理项目名中的非法字符，避免目录创建失败。
		project := createNameReg.ReplaceAllString(strings.TrimSpace(createProjectName), "")
		if project == "" {
			return nil, fmt.Errorf("project name is required")
		}

		// 规范化语言名，保持和模板缓存 key 一致。
		language := strings.ToLower(strings.TrimSpace(createLanguage))
		if language == "" {
			language = "go"
		}

		// 返回和交互式表单相同的数据结构，简化后续 create 流程。
		return &view.CreateFormEntity{
			Project:  project,
			Language: language,
		}, nil
	}

	// 未传入非交互参数时，沿用原有 Bubble Tea 表单。
	return view.NewCreate()
}
