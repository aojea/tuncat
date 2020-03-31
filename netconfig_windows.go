package main

import (
	"fmt"
	"os/exec"
	"strings"
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
	sargs := fmt.Sprintf("interface ip set address name=REPLACE_ME source=static addr=REPLACE_ME mask=REPLACE_ME gateway=none")
	args := strings.Split(sargs, " ")
	args[4] = fmt.Sprintf("name=%s", n.dev)
	args[6] = fmt.Sprintf("addr=%s", n.ip)
	// Set a /32 mask because the important is the route through the interface
	args[7] = fmt.Sprintf("mask=255.255.255.255")
	cmd := exec.Command("netsh", args...)
	return cmd.Run()
}

func (n Netconfig) CreateRoutes() error {
	// TODO
	return nil
}

func (n Netconfig) DeleteRoutes() error {
	// TODO
	return nil
}

func (n Netconfig) CreateMasquerade(dev string) error {
	// Only for Linux
	return nil
}

func (n Netconfig) DeleteMasquerade(dev string) error {
	// Only for Linux
	return nil
}
