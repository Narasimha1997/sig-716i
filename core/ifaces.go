package core

import (
	"log"
	"net"
	"strings"
)

type Wireless struct {
	Ifaces []net.Interface
}

func (w *Wireless) ProbeWirelessInterfaces() *InternalError {
	interfaces, err := net.Interfaces()

	if err != nil {
		return &InternalError{
			Status:  InterfaceProbeError,
			Message: err.Error(),
		}
	}

	for _, iface := range interfaces {
		ifaceName := iface.Name
		log.Printf("found interface name=%s, mac=%s", iface.Name, iface.HardwareAddr)
		if !strings.HasPrefix(ifaceName, IfacePrefixWifi) {
			log.Printf("skippig %s, not a wireless interface", ifaceName)
			continue
		}

		log.Printf("found wireless interface name=%s, mac=%s", ifaceName, iface.HardwareAddr)
		w.Ifaces = append(w.Ifaces, iface)
	}

	if len(w.Ifaces) == 0 {
		return &InternalError{
			Status:  InterfaceNoWireless,
			Message: "no wireless interfaces found",
		}
	}

	log.Printf("found %d wireless interfaces", len(w.Ifaces))
	return nil
}
