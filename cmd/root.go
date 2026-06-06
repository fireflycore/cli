package cmd

import (
	"github.com/fireflycore/cli/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

// templateVersion 保存 create 命令使用的 go-layout 模板版本。
var templateVersion string

// rootCmd 是 firefly CLI 的根命令，所有子命令都挂载在它下面。
var rootCmd = &cobra.Command{
	Use:     "firefly",
	Short:   "Firefly Go microservice engineering helper.",
	Version: config.RELEASE,
}

// Execute 执行根命令，并在 Cobra 返回错误时使用非零状态码退出。
func Execute() {
	// 执行 Cobra 命令树。
	err := rootCmd.Execute()
	if err != nil {
		// Cobra 已负责打印错误，这里只负责给 shell 一个失败状态码。
		os.Exit(1)
	}
}

func init() {
	// 根命令当前不定义全局 flag；具体配置由各子命令按自己的职责声明。
}
