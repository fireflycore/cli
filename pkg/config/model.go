package config

type CacheConfigEntity struct {
	Root     string `json:"root" yaml:"root" mapstructure:"root"`
	Config   string `json:"config" yaml:"config" mapstructure:"config"`
	Template string `json:"template" yaml:"template" mapstructure:"template"`
}

type CoreEntity struct {
	Setup   string            `json:"setup" yaml:"setup" mapstructure:"setup"`
	Cache   CacheConfigEntity `json:"cache" yaml:"cache" mapstructure:"cache"`
	Version map[string]string `json:"version" yaml:"version" mapstructure:"version"`
}
