package buf

type GenConfigEntity struct {
	Version string          `json:"version" yaml:"version" mapstructure:"version"`
	Types   *TypesEntity    `json:"types" yaml:"types" mapstructure:"types"`
	Managed *ManagedEntity  `json:"managed" yaml:"managed" mapstructure:"managed"`
	Plugins []*PluginEntity `json:"plugins" yaml:"plugins" mapstructure:"plugins"`
}

type TypesEntity struct {
	Include []string `json:"include" yaml:"include" mapstructure:"include"`
}

type ManagedEntity struct {
	Enabled         bool           `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	GoPackagePrefix *PackagePrefix `json:"go_package_prefix" yaml:"go_package_prefix" mapstructure:"go_package_prefix"`
}

type PackagePrefix struct {
	Default string   `json:"default" yaml:"default" mapstructure:"default"`
	Except  []string `json:"except" yaml:"except" mapstructure:"except"`
}

type PluginEntity struct {
	Plugin string `json:"plugin" yaml:"plugin" mapstructure:"plugin"`
	Out    string `json:"out" yaml:"out" mapstructure:"out"`
	Opt    string `json:"opt" yaml:"opt" mapstructure:"opt"`
}
