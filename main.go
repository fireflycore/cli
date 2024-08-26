/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"github.com/fireflycore/cli/pkg"
)

var ask = []string{
	"Please input your project name.",
	"Please choose your development language.",
}

func main() {
	//cmd.Execute()
	//ip := pkg.NewInput(ask[0])
	//if _, err := ip.Run(); err != nil {
	//	fmt.Println(err)
	//	return
	//}

	sp := pkg.NewSelect(ask[1], pkg.LANGUAGE)
	if _, err := sp.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
