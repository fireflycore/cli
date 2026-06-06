package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// New 初始化 CLI 配置上下文。
func New() (*CoreEntity, error) {
	// 先创建空配置对象，后续逐步填充路径和全局配置。
	core := CoreEntity{
		Global: &GlobalPersistenceStorageConfigEntity{},
	}

	// 获取用户级缓存目录，用来保存模板仓库和 CLI 配置。
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	// 获取当前 CLI 可执行文件路径。
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// 获取用户执行命令时所在的工作目录。
	cur, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 记录 CLI 可执行文件路径。
	core.SetupDir = exe
	// 记录当前工作目录，create 命令会把项目生成到这里。
	core.LocalDir = cur

	// 组织 CLI 根缓存目录。
	core.CacheDir = filepath.Join(cache, "cache", CLI_NAME)
	// 组织模板缓存目录。
	core.CacheTemplateDir = filepath.Join(core.CacheDir, "template")

	// 计算全局配置文件名。
	core.GlobalConfigFileName = fmt.Sprintf("%s.%s", CLI_CONFIG_FILE_NAME, CLI_CONFIG_FILE_TYPE)
	// 计算全局配置目录。
	core.GlobalConfigFilePath = filepath.Join(core.CacheDir, "config")

	// 加载或创建全局配置。
	if err = core.loadGlobalConfig(); err != nil {
		return nil, err
	}

	// 返回完整配置上下文。
	return &core, nil
}

// loadGlobalConfig 加载用户缓存目录中的全局配置，不存在时自动创建默认配置。
func (core *CoreEntity) loadGlobalConfig() error {
	// 创建独立 viper 实例，避免污染其他读取流程。
	core.gv = viper.New()
	// 设置配置文件名。
	core.gv.SetConfigName(CLI_CONFIG_FILE_NAME)
	// 设置配置格式。
	core.gv.SetConfigType(CLI_CONFIG_FILE_TYPE)
	// 设置配置目录。
	core.gv.AddConfigPath(core.GlobalConfigFilePath)

	// 检查全局配置文件是否存在。
	_, err := os.Stat(filepath.Join(core.GlobalConfigFilePath, core.GlobalConfigFileName))
	if err != nil {
		// 配置不存在时初始化默认配置。
		if os.IsNotExist(err) {
			// 确保全局配置目录存在。
			if err = os.MkdirAll(core.GlobalConfigFilePath, 0755); err != nil {
				return err
			}

			// 为每种支持语言写入默认 latest 模板版本。
			core.Global.Version = make(map[string]string)
			for _, language := range LANGUAGE {
				core.Global.Version[strings.ToLower(language)] = "latest"
			}

			// 将默认配置写入磁盘。
			if err = core.UpdateGlobalConfig(); err != nil {
				return err
			}
		} else {
			// 其他 stat 错误直接返回。
			return err
		}
	}

	// 从磁盘读取配置内容。
	if err = core.gv.ReadInConfig(); err != nil {
		return err
	}

	// 将配置反序列化到 Global 结构体。
	if err = core.gv.Unmarshal(&core.Global); err != nil {
		return err
	}

	return nil
}

// UpdateGlobalConfig 将当前全局配置写回用户缓存目录。
func (core *CoreEntity) UpdateGlobalConfig() error {
	// 将模板版本 map 写入 viper。
	core.gv.Set("version", core.Global.Version)

	// 使用 WriteConfigAs 明确写入目标文件。
	if err := core.gv.WriteConfigAs(filepath.Join(core.GlobalConfigFilePath, core.GlobalConfigFileName)); err != nil {
		return err
	}

	return nil
}
