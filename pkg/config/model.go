package config

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
}

type GlobalPersistenceStorageConfigEntity struct {
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}

type LocalPersistenceStorageConfigEntity struct {
	Language string `json:"language" yaml:"language" mapstructure:"language"`
	Version  string `json:"version" yaml:"version" mapstructure:"version"`
}
