package config

import (
	"github.com/spf13/viper"
)

type CoreEntity struct {
	SetupDir string

	LocalDir         string
	CacheDir         string
	CacheTemplateDir string

	Global               *GlobalPersistenceStorageConfigEntity
	GlobalConfigFileName string
	GlobalConfigFilePath string

	gv *viper.Viper
}

type GlobalPersistenceStorageConfigEntity struct {
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}
