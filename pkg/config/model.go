package config

type CacheConfigEntity struct {
	Root     string `json:"root" yaml:"root" mapstructure:"root"`
	Config   string `json:"config" yaml:"config" mapstructure:"config"`
	Template string `json:"template" yaml:"template" mapstructure:"template"`
}

type CoreEntity struct {
	Setup   string            `json:"setup" yaml:"setup" mapstructure:"setup"`
	Current string            `json:"current" yaml:"current" mapstructure:"current"`
	Cache   CacheConfigEntity `json:"cache" yaml:"cache" mapstructure:"cache"`
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}

type GlobalConfigEntity struct {
	SetupDir string

	LocalDir            string
	LocalConfigFilePath string

	CacheDir            string
	CacheTemplateDir    string
	CacheConfigFilePath string

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
