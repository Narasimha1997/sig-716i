package main

import (
	"fmt"
	"os"
	"os/signal"
	"sig-716i/core"
	"syscall"
)

func cleanUp(iface core.Wireless) {
	iface.RollbackHost(iface.SelectedIface.Name)
}

func main() {
	iface := core.Wireless{}

	args := core.ParseArgs()

	if !args.Revert {
		err := iface.PrepareHost(args.Iface)

		// prepare rollback on exit
		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChannel
			cleanUp(iface)
			os.Exit(0)
		}()

		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			os.Exit(err.Status)
		}

		selectedIface := iface.GetIface()
		err = core.ListenForPacketsOnIface(&selectedIface)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			cleanUp(iface)
			os.Exit(err.Status)
		}

	} else {
		err := iface.RollbackHost(args.Iface)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			cleanUp(iface)
			os.Exit(err.Status)
		}
	}
}
