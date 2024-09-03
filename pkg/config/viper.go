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
		Local:  &LocalPersistenceStorageConfigEntity{},
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

	core.GlobalConfigPath = filepath.Join(core.CacheDir, "config")
	core.LocalConfigPath = filepath.Join(core.LocalDir, "cmd", CLI_NAME)

	configFileName := fmt.Sprintf("%s.%s", CLI_CONFIG_FILE_NAME, CLI_CONFIG_FILE_TYPE)

	if err = core.loadGlobalConfig(configFileName); err != nil {
		return nil, err
	}

	if err = core.loadLocalConfig(configFileName); err != nil {
		return nil, err
	}

	return &core, nil
}

func (core *CoreEntity) loadGlobalConfig(file string) error {
	gv := viper.New()
	gv.SetConfigName("cli")
	gv.SetConfigType("yaml")
	gv.AddConfigPath(core.GlobalConfigPath)

	_, err := os.Stat(filepath.Join(core.GlobalConfigPath, file))
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(core.GlobalConfigPath, 0755); err != nil {
				return err
			}

			core.Global.Version = make(map[string]string)
			for _, language := range Language {
				core.Global.Version[strings.ToLower(language)] = "latest"
			}
			gv.Set("version", core.Global.Version)

			if err = gv.WriteConfigAs(filepath.Join(core.GlobalConfigPath, file)); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err = gv.ReadInConfig(); err != nil {
		return err
	}

	if err = gv.Unmarshal(&core.Global); err != nil {
		return err
	}

	return nil
}

func (core *CoreEntity) loadLocalConfig(file string) error {
	_, err := os.Stat(filepath.Join(core.LocalConfigPath, file))
	if err != nil {
		core.Local = nil
	} else {
		lv := viper.New()
		lv.SetConfigName("cli")
		lv.SetConfigType("yaml")
		lv.AddConfigPath(core.LocalConfigPath)

		if err = lv.ReadInConfig(); err != nil {
			return err
		}

		if err = lv.Unmarshal(&core.Local); err != nil {
			return err
		}
	}
	return nil
}
