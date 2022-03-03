package core

import (
	"log"
	"os"
	"os/exec"

	"github.com/akamensky/argparse"
)

func ExecCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.Output()
	if err != nil {
		return stdout, err
	}

	return stdout, nil
}

// populate and return the CLIArgs struct
func ParseArgs() *CLIArgs {
	parser := argparse.NewParser("brute", "a CLI tool for bombing wireless networks")
	revert := parser.Flag("r", "revert", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "revert back the host to normal mode of operation",
	})

	manualIface := parser.String("if", "iface", &argparse.Options{
		Required: false,
		Default:  "",
		Help:     "specify the host interface manually",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatalf("failed to parse arguments, invalid args")
	}

	return &CLIArgs{
		Revert: *revert,
		Iface:  *manualIface,
	}
}
