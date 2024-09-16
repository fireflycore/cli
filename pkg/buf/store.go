package buf

import "fmt"

func (gen *GenConfigEntity) AddGenStore(mode, store string) error {
	switch mode {
	case "module":
		for _, item := range gen.Inputs {
			if v, ok := item.(ModuleInputEntity); ok {
				if v.Module == store {
					return fmt.Errorf("store already exists")
				}
			}
		}
		gen.Inputs = append(gen.Inputs, ModuleInputEntity{
			Module: store,
		})
	case "local":
		for _, item := range gen.Inputs {
			if v, ok := item.(LocalInputEntity); ok {
				if v.Directory == store {
					return fmt.Errorf("store already exists")
				}
			}
		}
		gen.Inputs = append(gen.Inputs, LocalInputEntity{
			Directory: store,
		})
	}
	return nil
}

func (gen *GenConfigEntity) RemoveGenStore(store string) {
	fmt.Println(store)
}
