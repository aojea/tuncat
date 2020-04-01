package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func validate(ifAddress, remoteNetwork, remoteGateway string) error {
	// IP address of the local tun interface
	if net.ParseIP(ifAddress) == nil {
		return fmt.Errorf("Invalid Interface IP address")
	}

	// Remote network via the remote tunnel
	if remoteNetwork != "" {
		_, ipNet, err := net.ParseCIDR(remoteNetwork)
		if err != nil {
			return err
		}
		remoteNetwork = ipNet.Network()
	}

	// Remote gateway via the remote tunnel
	if len(remoteGateway) > 0 && net.ParseIP(remoteGateway) == nil {
		return fmt.Errorf("Invalid Remote Gateway IP address")
	}
	return nil
}

func main() {

	var remoteNetwork, remoteGateway, ifAddress string
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	remoteAddress := connectCmd.String("dst-host", "", "remote host address")
	remotePort := connectCmd.Int("dst-port", 0, "specify the local port to be used")
	connectCmd.StringVar(&ifAddress, "if-address", "192.168.166.1", "Local interface address")
	connectCmd.StringVar(&remoteNetwork, "remote-network", "", "Remote network via the tunnel")
	connectCmd.StringVar(&remoteGateway, "remote-gateway", "", "Remote gateway via the tunnel")

	listenCmd := flag.NewFlagSet("listen", flag.ExitOnError)
	sourceAddress := listenCmd.String("src-host", "0.0.0.0", "specify the local address to be used")
	sourcePort := listenCmd.Int("src-port", 0, "specify the local port to be used")

	if len(os.Args) < 2 {
		fmt.Println("usage: tuncat [<args>] <command>")
		flag.PrintDefaults()
		fmt.Println("tuncat commands are: ")
		fmt.Println(" connect [<args>] Connect to a remote host")
		fmt.Println(" listen [<args>] Listen on a local port")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "listen":
		listenCmd.Parse(os.Args[2:])
	case "connect":
		connectCmd.Parse(os.Args[2:])
	default:
		fmt.Println("usage: tuncat [<args>] <command>")
		flag.PrintDefaults()
		fmt.Println("tuncat commands are: ")
		fmt.Println(" connect [<args>] Connect to a remote host")
		connectCmd.PrintDefaults()
		fmt.Println(" listen [<args>] Listen on a local port")
		listenCmd.PrintDefaults()
		os.Exit(1)
	}

	// Global configuration
	flag.Parse()

	// Connect command
	if connectCmd.Parsed() {
		// Obtain remote port and remote address
		if *remoteAddress == "" || *remotePort == 0 {
			connectCmd.PrintDefaults()
			os.Exit(1)
		}
		// Configure a new client
		remoteHost := net.JoinHostPort(*remoteAddress, strconv.Itoa(*remotePort))
		// Validate configuration
		if err := validate(ifAddress, remoteNetwork, remoteGateway); err != nil {
			log.Fatalf("Validation error %v", err)
			os.Exit(1)
		}
		client := NewClient(remoteHost)
		client.IfAddress = ifAddress
		client.RemoteNetwork = remoteNetwork
		client.RemoteGateway = remoteGateway
		// Connect to the server
		if err := client.Start(); err != nil {
			log.Printf("Client error: %v", err)
		}
		client.Close()
	}

	// Listen command
	if listenCmd.Parsed() {
		if *sourcePort == 0 {
			listenCmd.PrintDefaults()
			os.Exit(1)
		}

		// Configure a new Server
		listenAddress := net.JoinHostPort(*sourceAddress, strconv.Itoa(*sourcePort))
		server := NewServer(listenAddress)
		// Validate configuration
		if err := validate(ifAddress, "", ""); err != nil {
			log.Fatalf("Validation error %v", err)
			os.Exit(1)
		}
		if ifAddress != "" {
			server.IfAddress = ifAddress
		}
		// Listen
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
		server.Close()
	}
}
