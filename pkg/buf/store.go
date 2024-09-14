package buf

import "fmt"

func (gen *GenConfigEntity) AddGenStore(mode, store string) {
	fmt.Println(mode, store, "睡个好觉")
}

func (gen *GenConfigEntity) RemoveGenStore(store string) {
	fmt.Println(store)
}
