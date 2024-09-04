package main

import (
	"fmt"
	"github.com/fireflycore/cli/cmd"
	"github.com/fireflycore/cli/pkg/config"
	"github.com/fireflycore/cli/pkg/store"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	store.Use.Config = cfg
	cmd.Execute()
}
