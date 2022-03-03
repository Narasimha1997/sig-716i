package main

import (
	"brute/core"
	"fmt"
	"os"
)

func main() {
	iface := core.Wireless{}

	args := core.ParseArgs()

	if !args.Revert {
		err := iface.PrepareHost(args.Iface)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			os.Exit(err.Status)
		}
	} else {
		err := iface.RollbackHost(args.Iface)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			os.Exit(err.Status)
		}
	}
}
