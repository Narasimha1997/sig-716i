package core

import "fmt"

type InternalError struct {
	Message string
	Status  int
}

func (i *InternalError) Error() string {
	return fmt.Sprintf("%s: %d", i.Message, i.Status)
}

const (
	InterfaceProbeError int = -1
	InterfaceNoWireless int = -2
)

const (
	IfacePrefixWifi string = "wl"
)
