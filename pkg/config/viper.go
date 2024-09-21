package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

func New() (*CoreEntity, error) {
	core := CoreEntity{
		Global: &GlobalPersistenceStorageConfigEntity{},
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	cur, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	core.SetupDir = exe
	core.LocalDir = cur

	core.CacheDir = filepath.Join(cache, "cache", CLI_NAME)
	core.CacheTemplateDir = filepath.Join(core.CacheDir, "template")

	core.GlobalConfigFileName = fmt.Sprintf("%s.%s", CLI_CONFIG_FILE_NAME, CLI_CONFIG_FILE_TYPE)
	core.GlobalConfigFilePath = filepath.Join(core.CacheDir, "config")

	if err = core.loadGlobalConfig(); err != nil {
		return nil, err
	}

	return &core, nil
}

func (core *CoreEntity) loadGlobalConfig() error {
	core.gv = viper.New()
	core.gv.SetConfigName(CLI_CONFIG_FILE_NAME)
	core.gv.SetConfigType(CLI_CONFIG_FILE_TYPE)
	core.gv.AddConfigPath(core.GlobalConfigFilePath)

	_, err := os.Stat(filepath.Join(core.GlobalConfigFilePath, core.GlobalConfigFileName))
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(core.GlobalConfigFilePath, 0755); err != nil {
				return err
			}

			core.Global.Version = make(map[string]string)
			for _, language := range LANGUAGE {
				core.Global.Version[strings.ToLower(language)] = "latest"
			}

			if err = core.UpdateGlobalConfig(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err = core.gv.ReadInConfig(); err != nil {
		return err
	}

	if err = core.gv.Unmarshal(&core.Global); err != nil {
		return err
	}

	return nil
}

func (core *CoreEntity) UpdateGlobalConfig() error {
	core.gv.Set("version", core.Global.Version)

	if err := core.gv.WriteConfigAs(filepath.Join(core.GlobalConfigFilePath, core.GlobalConfigFileName)); err != nil {
		return err
	}

	return nil
}
