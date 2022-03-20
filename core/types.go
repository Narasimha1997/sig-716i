package core

import "fmt"

type CLIArgs struct {
	Revert          bool
	Iface           string
	FilteredAPs     string
	FilteredClients string
}

type InternalError struct {
	Message string
	Status  int
}

func (i *InternalError) Error() string {
	return fmt.Sprintf("%s: %d", i.Message, i.Status)
}

const (
	InterfaceProbeError   int = -1
	InterfaceNoWireless   int = -2
	InterfaceCommandError int = -3
	InterfaceNotFound     int = -4
	InterfaceNoScanned    int = -5
	InterfaceMonModeError int = -6
	InterfaceNoRollback   int = -7
	ArgParseError         int = -8
	PcapHandleError       int = -9
)

const (
	IfacePrefixWifi string = "wl"
)
