package core

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Wireless struct {
	SelectedIface net.Interface
}

func (w *Wireless) probeWirelessInterfaces() ([]net.Interface, *InternalError) {
	interfaces, err := net.Interfaces()

	wirelessIfaces := make([]net.Interface, 0)

	if err != nil {
		return nil, &InternalError{
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
		wirelessIfaces = append(wirelessIfaces, iface)
	}

	if len(wirelessIfaces) == 0 {
		return nil, &InternalError{
			Status:  InterfaceNoWireless,
			Message: "no wireless interfaces found",
		}
	}

	log.Printf("found %d wireless interfaces", len(wirelessIfaces))
	return wirelessIfaces, nil
}

func (w *Wireless) downNetworkManager() *InternalError {
	_, err := ExecCommand("systemctl", "stop", "NetworkManager")
	if err != nil {
		return &InternalError{
			Status:  InterfaceCommandError,
			Message: fmt.Sprintf("failed to bring down network manager %s", err.Error()),
		}
	}

	return nil
}

func (w *Wireless) upNetworkManager() *InternalError {
	_, err := ExecCommand("systemctl", "start", "NetworkManager")
	if err != nil {
		return &InternalError{
			Status:  InterfaceCommandError,
			Message: fmt.Sprintf("failed to bring down network manager %s", err.Error()),
		}
	}

	return nil
}

func (w *Wireless) selectInterface(target string, ifaces []net.Interface) *InternalError {

	if target != "" {
		// user has manually specified a target
		for _, iface := range ifaces {
			if iface.Name == target {
				// make this as the selected interface and return
				w.SelectedIface = iface
				log.Printf("selected %s as the host interface", iface.Name)
				return nil
			}
		}

		// no such interface
		return &InternalError{
			Status:  InterfaceNotFound,
			Message: fmt.Sprintf("no such interface %s found in the machine", target),
		}
	}

	currentMaxDiscoveries := 0
	currentMaxIfaceIndex := -1

	for idx, iface := range ifaces {
		op, err := ExecCommand("iwlist", iface.Name, "scan")
		if err != nil {
			return &InternalError{
				Status:  InterfaceCommandError,
				Message: fmt.Sprintf("scan error %s", err.Error()),
			}
		}

		opStr := string(op)
		discoveryCount := strings.Count(opStr, " - Address:")

		log.Printf("found %d devices connected to %s", discoveryCount, iface.Name)
		if discoveryCount >= currentMaxDiscoveries {
			currentMaxDiscoveries = discoveryCount
			currentMaxIfaceIndex = idx
		}
	}

	if currentMaxIfaceIndex == -1 {
		return &InternalError{
			Status:  InterfaceNoScanned,
			Message: "no interfaces had connected devices",
		}
	}

	selectedInterface := ifaces[currentMaxIfaceIndex]
	log.Printf("selected %s as the host interface", selectedInterface.Name)

	w.SelectedIface = selectedInterface
	return nil
}

func (w *Wireless) toggleMode(mode string) *InternalError {

	iface := w.SelectedIface.Name

	errReturn := func(err error) *InternalError {
		return &InternalError{
			Status:  InterfaceMonModeError,
			Message: fmt.Sprintf("failed to turn on monitoring mode for %s: %s", iface, err.Error()),
		}
	}

	_, err := ExecCommand("ip", "link", "set", iface, "down")
	if err != nil {
		return errReturn(err)
	}

	// log.Printf("%s\n", string(op))

	_, err = ExecCommand("iwconfig", iface, "mode", mode)
	if err != nil {
		return errReturn(err)
	}

	// log.Printf("%s\n", string(op))

	_, err = ExecCommand("ip", "link", "set", iface, "up")
	if err != nil {
		return errReturn(err)
	}

	// log.Printf("%s\n", string(op))

	return nil
}

// SetupBaseInterface prepares the host machine to launch the attack
func (w *Wireless) PrepareHost(target string) *InternalError {
	log.Println("scanning and selecting available network interfaces on the machine...")
	ifaces, ifErr := w.probeWirelessInterfaces()
	if ifErr != nil {
		return ifErr
	}

	log.Println("selecting the suitable network interface...")
	ifErr = w.selectInterface(target, ifaces)
	if ifErr != nil {
		return ifErr
	}

	log.Printf(
		"turning on monitoring mode for %s, mac=%s",
		w.SelectedIface.Name,
		w.SelectedIface.HardwareAddr.String(),
	)

	log.Println("bringing down network manager....")
	ifErr = w.downNetworkManager()
	if ifErr != nil {
		return ifErr
	}

	ifErr = w.toggleMode("monitor")
	if ifErr != nil {
		return ifErr
	}

	log.Println("prepared host interface")
	return nil
}

func (w *Wireless) RollbackHost(ifname string) *InternalError {
	log.Println("rolling back host configuration...")

	if ifname == "" {
		return &InternalError{
			Status:  InterfaceNoRollback,
			Message: "no interface specified to rollback",
		}
	}

	w.SelectedIface.Name = ifname

	ifErr := w.toggleMode("managed")
	if ifErr != nil {
		return ifErr
	}

	log.Println("starting network manager...")
	ifErr = w.upNetworkManager()
	if ifErr != nil {
		return ifErr
	}

	log.Println("rolled back host interface")
	return nil
}

func (w *Wireless) GetIface() net.Interface {
	return w.SelectedIface
}
