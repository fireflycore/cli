package main

import (
	"fmt"
	"github.com/fireflycore/cli/cmd"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/store"
)

func main() {
	// 初始化 CLI 全局配置，包括缓存目录、本地目录和全局版本配置。
	cfg, err := config.New()
	if err != nil {
		// 初始化失败时直接打印错误并退出，避免后续命令拿到空配置。
		fmt.Println(err)
		return
	}

	// 将配置放入全局 store，供各个命令和业务包读取。
	store.Use.Config = cfg
	// 进入 Cobra 根命令执行流程。
	cmd.Execute()
}
