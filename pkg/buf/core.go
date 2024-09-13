package buf

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type CoreEntity struct {
	Config *GenConfigEntity

	p string
	v *viper.Viper
}

func (core *CoreEntity) WriteConfig() error {
	if err := core.v.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func (core *CoreEntity) GetModule() []string {
	var list []string
	for _, item := range core.Config.Inputs {
		fmt.Println(item)
		if v, ok := item.(ModuleInputEntity); ok {
			list = append(list, v.Module)
		}
	}
	return list
}

func New(path string) (*CoreEntity, error) {
	core := &CoreEntity{
		Config: &GenConfigEntity{},
		p:      path,
	}

	core.v = viper.New()
	core.v.SetConfigName("buf.gen")
	core.v.SetConfigType("yaml")
	core.v.AddConfigPath(path)

	if _, err := os.Stat(filepath.Join(core.p, "buf.gen.yaml")); err != nil {
		return nil, err
	}
	if err := core.v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := core.v.Unmarshal(core.Config); err != nil {
		return nil, err
	}

	for ii, item := range core.Config.Inputs {
		if v, ok := item.(map[string]interface{}); ok {
			var row interface{}
			if _, ok = v["module"]; ok {
				row = ModuleInputEntity{}
			}
			if _, ok = v["directory"]; ok {
				row = LocalInputEntity{}
			}
			_ = mapstructure.Decode(item, &row)
			core.Config.Inputs[ii] = row
		}
	}

	return core, nil
}
