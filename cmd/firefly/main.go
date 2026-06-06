package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fireflycore/cli/internal/cli"
)

func main() {
	// 从当前进程环境构造 CLI 运行时上下文。
	runtime, err := cli.NewRuntime()
	if err != nil {
		// 运行时初始化失败时写入 stderr 并返回非零状态。
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 执行 Cobra 命令树。
	if err = cli.Execute(context.Background(), runtime, nil); err != nil {
		// Cobra 已经负责打印使用说明，这里补充错误并返回非零状态。
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
