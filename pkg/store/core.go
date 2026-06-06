package store

import (
	"github.com/fireflycore/cli/pkg/config"
)

// _CoreEntity 是 CLI 运行期共享状态的容器。
type _CoreEntity struct {
	// Config 保存全局配置、缓存目录和本地目录信息。
	Config *config.CoreEntity
}

// Use 是全局共享状态入口，由 main 初始化后供各命令读取。
var Use = new(_CoreEntity)
