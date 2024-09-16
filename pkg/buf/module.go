package buf

import "fmt"

func (gen *GenConfigEntity) AddGenModule(store, module string) error {
	for i, item := range gen.Inputs {
		if v, ok := item.(ModuleInputEntity); ok {
			if v.Module == store {
				for _, m := range v.Types {
					if m == module {
						return fmt.Errorf("module already exists")
					}
				}
				v.Types = append(v.Types, module)
				gen.Inputs[i] = v
			}
		}
	}
	return nil
}

func (gen *GenConfigEntity) RemoveGenModule(module string) {
	fmt.Println(module)
}
