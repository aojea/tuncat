package main

import (
	"os/exec"
)

// Route represent a route
type Route struct {
	network string
	gw      string
}

// Netconfig represent the network configuration of an interface
type Netconfig struct {
	ip string
	// TODO: []Route
	routes Route
	dev    string
}

// NewNetconfig create new network configuration
func NewNetconfig(ip, remoteNetwork, remoteGateway, dev string) Netconfig {
	return Netconfig{
		ip: ip,
		routes: Route{
			network: remoteNetwork,
			gw:      remoteGateway,
		},
		dev: dev,
	}
}

func (n Netconfig) SetupNetwork() error {
	if err := exec.Command("ifconfig", n.dev, "inet", n.ip, n.ip, "up").Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) CreateRoutes() error {
	if len(n.routes.network) == 0 {
		return nil
	}
	return exec.Command("route", "-n", "add", n.routes.network, n.routes.gw).Run()
}

func (n Netconfig) DeleteRoutes() error {
	if len(n.routes.network) == 0 {
		return nil
	}
	return exec.Command("route", "-n", "delete", n.routes.network, n.routes.gw).Run()
}

func (n Netconfig) CreateMasquerade(dev string) error {
	// Only for Linux
	return nil
}

func (n Netconfig) DeleteMasquerade(dev string) error {
	// Only for Linux
	return nil
}
