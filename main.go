package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
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

	if err := validate(ifAddress, remoteNetwork, remoteGateway); err != nil {
		log.Fatalf("Validation error %v", err)
		os.Exit(1)
	}

	// Connect command
	if connectCmd.Parsed() {
		// Obtain remote port and remote address
		if *remoteAddress == "" || *remotePort == 0 {
			connectCmd.PrintDefaults()
			os.Exit(1)
		}
		// Connect to the remote address
		remoteHost := net.JoinHostPort(*remoteAddress, strconv.Itoa(*remotePort))
		log.Printf("Connecting to %s", remoteHost)
		conn, err := net.Dial("tcp", remoteHost)
		if err != nil {
			log.Fatalf("Can't connect to server %q: %v", remoteHost, err)
		}
		defer conn.Close()
		// Establish the connection: send the tunnel parameters
		errChan := make(chan error, 1)
		timeout := 10 * time.Second
		go func() {
			errChan <- ClientConnect(conn, remoteNetwork, remoteGateway)
		}()

		// wait for the first thing to happen, either
		// an error, a timeout, or a result
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatalf("Can't establish connection: %v", err)
			}
		case <-time.After(timeout):
			log.Fatal("Can't establish connection: Timed Out")
		}
		// Create the Host Interface
		log.Println("Create Host Interface ...")
		ifce, err := NewHostInterface(ifAddress, remoteNetwork, remoteGateway, false)
		if err != nil {
			log.Fatalf("Error creating Host Interface: %v", err)
		}
		// Create the tunnel in client mode
		tun := NewTunnel(conn, ifce)
		// Run the tunnel until it fails or is killed
		tun.Run()
	}

	// Listen command
	if listenCmd.Parsed() {
		if *sourcePort == 0 {
			listenCmd.PrintDefaults()
			os.Exit(1)
		}
		// Listen
		sourceHost := net.JoinHostPort(*sourceAddress, strconv.Itoa(*sourcePort))
		ln, err := net.Listen("tcp", sourceHost)
		if err != nil {
			log.Fatalf("Can't Listen on address %s : %v", sourceHost, err)
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalf("Can't accept connection on address %s : %v", sourceHost, err)
			}
			// Establish the connection: receive the tunnel parameters
			errChan := make(chan error, 1)
			timeout := 10 * time.Second
			go func() {
				remoteNetwork, remoteGateway, err = ServerConnect(conn)
				errChan <- err
			}()
			// wait for the first thing to happen, either
			// an error, a timeout, or a result
			select {
			case err := <-errChan:
				if err != nil {
					// Closeon error and wait for a new connection
					log.Printf("Can't establish connection: %v", err)
					conn.Close()
					continue
				}
			case <-time.After(timeout):
				// Closeon error and wait for a new connection
				log.Printf("Can't establish connection: TimeOut")
				conn.Close()
				continue
			}

			// Create the Host Interface
			log.Println("Creating Host Interface ...")
			ifce, err := NewHostInterface(ifAddress, remoteNetwork, remoteGateway, true)
			if err != nil {
				log.Fatalf("Error creating Host Interface: %v", err)
			}
			// Create the tunnel and block (only accept one connection)
			tun := NewTunnel(conn, ifce)
			fmt.Println("Running the tunnel")
			tun.Run()
		}
	}

}
