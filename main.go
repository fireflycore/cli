package main

import (
	"fmt"
	"github.com/fireflycore/cli/cmd"
	"github.com/fireflycore/cli/pkg/buf"
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

	if cli, err := buf.New(store.Use.Config.LocalDir); err == nil {
		store.Use.Buf = cli
	}

	cmd.Execute()
}
