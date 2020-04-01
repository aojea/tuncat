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

// SetupNetwork configure the interface
func (n Netconfig) SetupNetwork() error {
	if err := exec.Command("ip", "link", "set", n.dev, "up").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "addr", "add", n.ip, "dev", n.dev).Run(); err != nil {
		return err
	}
	return nil
}

// CreateRoutes configure the routes associated to the interface
func (n Netconfig) CreateRoutes() error {
	if len(n.routes.network) == 0 {
		return nil
	}
	return exec.Command("ip", "route", "add", n.routes.network, "via", n.routes.gw).Run()
}

// DeleteRoutes deletes the routes associated to the interface
func (n Netconfig) DeleteRoutes() error {
	if len(n.routes.network) == 0 {
		return nil
	}
	return exec.Command("ip", "route", "del", n.routes.network, "via", n.routes.gw).Run()
}

// CreateMasquerade configures the network so the outgoing traffic is masquerade
// and the incoming traffic is sent through the tunnel using policy based source routing
func (n Netconfig) CreateMasquerade(dev string) error {
	if len(n.routes.network) == 0 {
		return nil
	}
	// Masquerade the tunnel traffic with the external interface
	if err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "route", "add", "table", "10", "to", "default", "via", n.ip).Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "rule", "add", "from", n.routes.network, "table", "10", "priority", "10").Run(); err != nil {
		return err
	}
	return nil
}

// DeleteMasquerade rules
func (n Netconfig) DeleteMasquerade(dev string) error {
	if len(n.routes.network) == 0 {
		return nil
	}
	if err := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run(); err != nil {
		return err
	}

	if err := exec.Command("ip", "route", "flush", "10").Run(); err != nil {
		return err
	}

	if err := exec.Command("ip", "rule", "del", "from", n.routes.network, "table", "10", "priority", "10").Run(); err != nil {
		return err
	}
	return nil
}
