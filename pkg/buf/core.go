package buf

import (
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

	return core, nil
}
