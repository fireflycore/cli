package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

func New() (*CoreEntity, error) {
	global := CoreEntity{}
	//var current string

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	global.Cache.Root = filepath.Join(cache, CLI_NAME)
	global.Cache.Config = filepath.Join(global.Cache.Root, "config")
	global.Cache.Template = filepath.Join(global.Cache.Root, "template")

	configFilePath := filepath.Join(global.Cache.Config, fmt.Sprintf("%s.%s", CLI_CONFIG_FILE_NAME, CLI_CONFIG_FILE_TYPE))

	v := viper.New()
	v.SetConfigName("cli")
	v.SetConfigType("yaml")
	v.AddConfigPath(global.Cache.Config)

	_, err = os.Stat(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(global.Cache.Config, 0755); err != nil {
				return nil, err
			}

			global.Version = make(map[string]string)
			for _, language := range Language {
				global.Version[language] = "latest"
			}

			v.Set("cache", global.Cache)
			v.Set("version", global.Version)

			if err = v.WriteConfigAs(configFilePath); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err = v.Unmarshal(&global); err != nil {
		return nil, err
	}

	return &global, nil
}
