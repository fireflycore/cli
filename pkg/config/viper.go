package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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

	v := viper.New()
	v.SetConfigName("cli")
	v.SetConfigType("yaml")
	v.AddConfigPath(core.GlobalConfigPath)

	_, err = os.Stat(filepath.Join(core.GlobalConfigPath, configFileName))
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(core.GlobalConfigPath, 0755); err != nil {
				return nil, err
			}

			core.Global.Version = make(map[string]string)
			for _, language := range Language {
				core.Global.Version[language] = "latest"
			}
			v.Set("version", core.Global.Version)

			if err = v.WriteConfigAs(filepath.Join(core.GlobalConfigPath, configFileName)); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err = v.Unmarshal(&core.Global); err != nil {
		return nil, err
	}

	return &core, nil
}
