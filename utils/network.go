package utils

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"net"
	"strings"
)

func GetDefaultDevice() (pcap.Interface, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return pcap.Interface{}, fmt.Errorf("error determining default IP address: %v", err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	defaultIP := localAddr.IP.String()

	devices, err := pcap.FindAllDevs()
	if err != nil {
		return pcap.Interface{}, fmt.Errorf("error finding devices: %v", err)
	}

	if len(devices) == 0 {
		return pcap.Interface{}, fmt.Errorf("no devices found")
	}

	var defaultDevice pcap.Interface
	for _, device := range devices {
		for _, address := range device.Addresses {
			if strings.Contains(address.IP.String(), defaultIP) {
				defaultDevice = device
				break
			}
		}
		if defaultDevice.Name != "" {
			break
		}
	}

	if defaultDevice.Name == "" {
		return pcap.Interface{}, fmt.Errorf("no suitable default device found")
	}

	return defaultDevice, nil
}

func CheckPcapInstalled() bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return false
	}

	if len(devices) == 0 {
		return false
	}

	return true
}

func FindDevice(device string) (pcap.Interface, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return pcap.Interface{}, fmt.Errorf("error finding devices: %v", err)
	}

	for _, d := range devices {
		if d.Description == device {
			return d, nil
		}
	}

	return pcap.Interface{}, fmt.Errorf("device not found")
}
