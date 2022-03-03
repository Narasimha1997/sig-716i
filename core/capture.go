package core

import (
	"fmt"
	"log"
	"net"
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
	BSSID   string
	Channel byte
}

type APManager struct {
	APMutex sync.Mutex
	APMap   map[string]AP

	ClientMutex sync.Mutex
	ClientsList []string

	DefaultChannel byte
}

func NewAPManager() *APManager {
	return &APManager{
		APMutex:        sync.Mutex{},
		APMap:          make(map[string]AP),
		DefaultChannel: 6,
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

func (a *APManager) addAPList(bssid string, packet gopacket.Packet) {

	a.APMutex.Lock()
	defer a.APMutex.Unlock()

	// BSSID already in map?
	_, ok := a.APMap[bssid]
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

	log.Printf("adding new AP with BSSID: %s, SSID: %s on channel: %d", ap.BSSID, ap.SSID, ap.Channel)
	a.APMap[bssid] = ap
}

func (a *APManager) CheckAndParseDot11(packet gopacket.Packet) {
	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer != nil {
		// we found a dot11 packet
		dot11Packet := dot11Layer.(*layers.Dot11)
		addr1 := strings.ToLower(dot11Packet.Address1.String())
		addr2 := strings.ToLower(dot11Packet.Address2.String())

		// is this access point already in the list?
		// the probe response originate from AP as a response to client probe
		// becaons originate periodically from APs used for discovery purposes
		dot11Probe := packet.Layer(layers.LayerTypeDot11MgmtProbeResp)
		dot11Beacon := packet.Layer(layers.LayerTypeDot11MgmtBeacon)

		if dot11Probe != nil || dot11Beacon != nil {
			// get all necessary details:
			addr3 := strings.ToLower(dot11Packet.Address3.String())
			a.addAPList(addr3, packet)
		}

		if a.isNoise(addr1, addr2) {
			return
		}

		// management packets, these packets will be used by clients to communicate
		// management related information with APs
		// log.Printf("found dot11 packet: %v", dot11Packet)
	}
}

var apManager *APManager = nil

func ListenForPacketsOnIface(iface *net.Interface) *InternalError {
	handle, err := pcap.OpenLive(iface.Name, 4096, false, 30*time.Second)
	if err != nil {
		return &InternalError{
			Status:  PcapHandleError,
			Message: fmt.Sprintf("failed to open pcacp handle: %s", err.Error()),
		}
	}

	apManager = NewAPManager()
	log.Println("initialized Aaccess points manager")

	pcapSource := gopacket.NewPacketSource(handle, handle.LinkType())

	log.Printf("created pcap source, waiting for Dot11 packets on %s", iface.Name)
	for packet := range pcapSource.Packets() {
		apManager.CheckAndParseDot11(packet)
	}

	return nil
}
