package main

import (
	"log"
	"runtime"

	"github.com/songgao/water"
)

const defaultInterface = "eth0"

// HostInterface represents the TUN interface and the networking configuration associated
type HostInterface struct {
	ifce       *water.Interface
	netCfg     Netconfig
	serverMode bool
}

// NewHostInterface returns a new HostInterface
func NewHostInterface(ifAddress, remoteNetwork, remoteGateway string, serverMode bool) (HostInterface, error) {
	// Create TUN interface
	// TODO: Windows have some network specific parameters
	// https://github.com/songgao/water/blob/master/params_windows.go
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return HostInterface{}, err
	}
	log.Printf("Interface Name: %s\n", ifce.Name())

	// Create the networking configuration
	// Set up routes to remote network depending if we are a server or a client
	var gw string
	if serverMode {
		gw = remoteGateway
	} else {
		gw = ifAddress
	}
	netCfg := NewNetconfig(ifAddress, remoteNetwork, gw, ifce.Name())
	// The network configuration is deleted when the interface is destroyed
	if err := netCfg.SetupNetwork(); err != nil {
		return HostInterface{}, err
	}

	log.Printf("Interface Up: %s\n", ifce.Name())

	log.Printf("Add route %v\n", netCfg.routes)
	if err := netCfg.CreateRoutes(); err != nil {
		return HostInterface{}, err
	}
	// Masquerade traffic in server mode and Linux
	if serverMode && runtime.GOOS == "linux" {
		log.Printf("Add Masquerade on interface %s\n", defaultInterface)
		if err := netCfg.CreateMasquerade(defaultInterface); err != nil {
			return HostInterface{}, err
		}
	}

	return HostInterface{
		ifce:       ifce,
		netCfg:     netCfg,
		serverMode: serverMode,
	}, nil

}

// Delete the interface created and the networking configuration associated
func (h HostInterface) Delete() {
	log.Printf("Delete interface routes: %q", h.ifce.Name())
	if err := h.netCfg.DeleteRoutes(); err != nil {
		log.Printf("Error deleting routes: %v", err)
	}
	// Masquerade traffic in server mode and Linux
	if h.serverMode && runtime.GOOS == "linux" {
		log.Printf("Delete interface routes: %q", h.ifce.Name())
		if err := h.netCfg.DeleteMasquerade(defaultInterface); err != nil {
			log.Printf("Error deleting masquerade: %v", err)
		}
	}
	h.ifce.Close()
}
