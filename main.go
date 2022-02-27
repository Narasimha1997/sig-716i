package main

import (
	"brute/core"
	"fmt"
	"os"
)

func main() {
	iface := core.Wireless{}
	err := iface.ProbeWirelessInterfaces()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(err.Status)
	}
}
