package main

import (
	"fmt"
	"github.com/fireflycore/cli/pkg/config"
)

func main() {
	//cmd.Execute()
	entity, err := config.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(entity)
}
