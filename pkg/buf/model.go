package buf

type GenConfigEntity struct {
	Version string         `json:"version" yaml:"version" mapstructure:"version"`
	Managed *ManagedEntity `json:"managed" yaml:"managed" mapstructure:"managed"`
	Plugins []interface{}  `json:"plugins" yaml:"plugins" mapstructure:"plugins"`
	Inputs  []interface{}  `json:"inputs" yaml:"inputs" mapstructure:"inputs"`
}

type ManagedEntity struct {
	Enabled  bool              `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Disable  []*DisableEntity  `json:"disable" yaml:"disable" mapstructure:"disable"`
	Override []*OverrideEntity `json:"override" yaml:"override" mapstructure:"override"`
}

type DisableEntity struct {
	FileOption string `json:"file_option" yaml:"file_option" mapstructure:"file_option"`
	Module     string `json:"module" yaml:"module" mapstructure:"module"`
}

type OverrideEntity struct {
	FileOption string `json:"file_option" yaml:"file_option" mapstructure:"file_option"`
	Value      string `json:"value" yaml:"value" mapstructure:"value"`
}

type LocalPluginEntity struct {
	Local string `json:"local" yaml:"local" mapstructure:"local"`
	Out   string `json:"out" yaml:"out" mapstructure:"out"`
	Opt   string `json:"opt" yaml:"opt" mapstructure:"opt"`
}

type RemotePluginEntity struct {
	Remote string `json:"remote" yaml:"remote" mapstructure:"remote"`
	Out    string `json:"out" yaml:"out" mapstructure:"out"`
	Opt    string `json:"opt" yaml:"opt" mapstructure:"opt"`
}

type ModuleInputEntity struct {
	Module string   `json:"module" yaml:"module" mapstructure:"module"`
	Types  []string `json:"types" yaml:"types" mapstructure:"types"`
}

type LocalInputEntity struct {
	Directory string `json:"directory" yaml:"directory" mapstructure:"directory"`
}
