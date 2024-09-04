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

	core.ConfigFileName = fmt.Sprintf("%s.%s", CLI_CONFIG_FILE_NAME, CLI_CONFIG_FILE_TYPE)

	if err = core.loadGlobalConfig(); err != nil {
		return nil, err
	}

	if err = core.loadLocalConfig(); err != nil {
		return nil, err
	}

	return &core, nil
}

func (core *CoreEntity) loadGlobalConfig() error {
	core.gv = viper.New()
	core.gv.SetConfigName("cli")
	core.gv.SetConfigType("yaml")
	core.gv.AddConfigPath(core.GlobalConfigPath)

	_, err := os.Stat(filepath.Join(core.GlobalConfigPath, core.ConfigFileName))
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(core.GlobalConfigPath, 0755); err != nil {
				return err
			}

			core.Global.Version = make(map[string]string)
			for _, language := range Language {
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

	if err := core.gv.WriteConfigAs(filepath.Join(core.GlobalConfigPath, core.ConfigFileName)); err != nil {
		return err
	}

	return nil
}

func (core *CoreEntity) loadLocalConfig() error {
	_, err := os.Stat(filepath.Join(core.LocalConfigPath, core.ConfigFileName))
	if err != nil {
		core.Local = nil
	} else {
		core.lv = viper.New()
		core.lv.SetConfigName("cli")
		core.lv.SetConfigType("yaml")
		core.lv.AddConfigPath(core.LocalConfigPath)

		if err = core.lv.ReadInConfig(); err != nil {
			return err
		}

		if err = core.lv.Unmarshal(&core.Local); err != nil {
			return err
		}
	}
	return nil
}

func (core *CoreEntity) UpdateLocalConfig() error {
	core.lv.Set("language", core.Local.Language)
	core.lv.Set("version", core.Local.Version)

	if err := core.lv.WriteConfigAs(filepath.Join(core.LocalConfigPath, core.ConfigFileName)); err != nil {
		return err
	}

	return nil
}
