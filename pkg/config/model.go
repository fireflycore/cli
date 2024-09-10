package config

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type CoreEntity struct {
	SetupDir string

	LocalDir        string
	LocalConfigPath string

	ConfigFileName string

	CacheDir         string
	CacheTemplateDir string

	GlobalConfigPath string

	Global *GlobalPersistenceStorageConfigEntity
	Local  *LocalPersistenceStorageConfigEntity

	gv *viper.Viper
	lv *viper.Viper
}

type GlobalPersistenceStorageConfigEntity struct {
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}

type LocalPersistenceStorageConfigEntity struct {
	Language string `json:"language" yaml:"language" mapstructure:"language"`
	Version  string `json:"version" yaml:"version" mapstructure:"version"`
}

type ColorConfig struct {
	Primary lipgloss.Style
	Warning lipgloss.Style
	Danger  lipgloss.Style
	Info    lipgloss.Style
	Focus   lipgloss.Style
}
