package buf

import "fmt"

func (gen *GenConfigEntity) AddGenStore(mode, store string) {
	fmt.Println(mode, store)
}

func (gen *GenConfigEntity) RemoveGenStore(store string) {
	fmt.Println(store)
}
