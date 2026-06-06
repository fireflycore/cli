package config

import (
	"github.com/spf13/viper"
)

// CoreEntity 保存 CLI 启动后需要共享的路径和全局配置。
type CoreEntity struct {
	// SetupDir 是当前 CLI 可执行文件路径。
	SetupDir string

	// LocalDir 是用户执行 firefly 命令时所在的当前目录。
	LocalDir string
	// CacheDir 是 CLI 的根缓存目录。
	CacheDir string
	// CacheTemplateDir 是模板仓库缓存目录。
	CacheTemplateDir string

	// Global 保存全局持久化配置内容。
	Global *GlobalPersistenceStorageConfigEntity
	// GlobalConfigFileName 是全局配置文件名。
	GlobalConfigFileName string
	// GlobalConfigFilePath 是全局配置文件所在目录。
	GlobalConfigFilePath string

	// gv 是读取和写入全局配置的 viper 实例。
	gv *viper.Viper
}

// GlobalPersistenceStorageConfigEntity 是 CLI 写入用户缓存目录的全局配置结构。
type GlobalPersistenceStorageConfigEntity struct {
	// Version 保存不同语言模板的默认版本，例如 go -> latest。
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}
