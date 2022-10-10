package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/akamensky/argparse"
)

func ExecCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s - %s", string(stdout), err.Error())
		return stdout, err
	}

	return stdout, nil
}

// populate and return the CLIArgs struct
func ParseArgs() *CLIArgs {
	parser := argparse.NewParser("sig-716i", "a CLI tool for bombing wireless networks")
	revert := parser.Flag("r", "revert", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "revert back the host to normal mode of operation",
	})

	manualIface := parser.String("i", "iface", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "specify the host interface manually",
	})

	targetAPs := parser.String("a", "aps", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "list of target BSSIDs of Access Points, ex: 52:54:00:eb:16:9d,02:42:93:53:b4:7b,...",
	})

	targetClients := parser.String("c", "clients", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "list of MAC addresses of the client devices, ex: 52:54:00:eb:16:9d,02:42:93:53:b4:7b,...",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatalf("failed to parse arguments, invalid args")
	}

	return &CLIArgs{
		Revert:          *revert,
		Iface:           *manualIface,
		FilteredAPs:     strings.ToLower(*targetAPs),
		FilteredClients: strings.ToLower(*targetClients),
	}
}
