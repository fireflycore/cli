package buf

import "fmt"

func (gen *GenConfigEntity) AddGenModule(store, module string) {
	fmt.Println(module, store)
}

func (gen *GenConfigEntity) RemoveGenModule(module string) {
	fmt.Println(module)
}
