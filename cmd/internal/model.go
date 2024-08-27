package internal

type ConfigEntity struct {
	Project  string            `json:"project" yaml:"project" mapstructure:"project"`
	Language string            `json:"language" yaml:"language" mapstructure:"language"`
	Database []*DatabaseEntity `json:"database" yaml:"database" mapstructure:"database"`
}

type DatabaseEntity struct {
	Type   string `json:"type" yaml:"type" mapstructure:"type"`
	Name   string `json:"name" yaml:"name" mapstructure:"name"`
	Url    string `json:"url" yaml:"url" mapstructure:"url"`
	Select bool   `json:"select" yaml:"select" mapstructure:"select"`
}
