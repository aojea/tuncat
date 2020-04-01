package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/songgao/water"
)

const defaultInterface = "eth0"

// Server represents a server instance.
type Server struct {
	conn   net.Conn
	ifce   *water.Interface
	netCfg Netconfig
	// Config
	IfAddress     string
	ListenAddress string
	//
	remoteNetwork string
	remoteGateway string
}

// NewServer returns a new instance of Server with default settings.
func NewServer(listenAddress string) *Server {
	return &Server{
		ListenAddress: listenAddress,
		// Configure one that doesn't overlap
		IfAddress: "192.168.166.1",
	}
}

// Start a new tunnel client
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		log.Fatalf("Can't Listen on address %s : %v", s.ListenAddress, err)
	}

	for {
		s.conn, err = ln.Accept()
		if err != nil {
			log.Fatalf("Can't accept connection on address %s : %v", s.ListenAddress, err)
		}
		// Establish the connection: receive the tunnel parameters
		errChan := make(chan error, 1)
		timeout := 10 * time.Second
		go func() {
			errChan <- s.handShake()
		}()
		// wait for the first thing to happen, either
		// an error, a timeout, or a result
		select {
		case err := <-errChan:
			if err != nil {
				// Closeon error and wait for a new connection
				log.Printf("Can't establish connection: %v", err)
				s.conn.Close()
				continue
			}
		case <-time.After(timeout):
			// Closeon error and wait for a new connection
			log.Printf("Can't establish connection: TimeOut")
			s.conn.Close()
			continue
		}

		// Create the Host Interface
		log.Println("Create Host Interface ...")
		err = s.createInterface()
		if err != nil {
			return fmt.Errorf("Error creating Host Interface: %v", err)
		}
		// Configure the interface network
		log.Println("Setup Interface Network...")
		err = s.setupNetwork()
		if err != nil {
			return fmt.Errorf("Error creating Host Interface: %v", err)
		}
		// Run the tunnel and block we only accept one connection
		Tunnel(s.conn, s.ifce)
		s.Close()
	}
}

// Close disconnects the underlying connection to the server.
func (s *Server) Close() {
	dev := defaultInterface
	log.Println("Shutting down the server...")
	// Close the connection
	if s.conn != nil {
		s.conn.Close()
	}
	// Delete host interface network configuration
	if err := s.netCfg.DeleteRoutes(); err != nil {
		log.Printf("Error deleting routes: %v", err)
	}
	// Delete host interface network configuration
	if err := s.netCfg.DeleteMasquerade(dev); err != nil {
		log.Printf("Error deleting masquerade rules: %v", err)
	}
	// Close interface
	if s.ifce != nil {
		s.ifce.Close()
	}
}

// handShake do the tunnel connection negotiation sending the configuration parameters for the serveer
// messages has to be echoed from the server in order to the connection to be established
func (s *Server) handShake() error {
	reader := bufio.NewReader(s.conn)
	// Receive the configuration parameters
	// will listen for message to process ending in newline (\n)
	message, _ := reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	// process for string received, we should receive the remoteNetwork parameter
	m := strings.Split(message, ":")
	if m[0] != "remoteNetwork" {
		return fmt.Errorf("Connection error, Received: %s Expected: remoteNetwork", m[0])
	}
	s.remoteNetwork = strings.TrimSpace(m[1])
	// send string back to client for ACK
	s.conn.Write([]byte(message))
	// will listen for message to process ending in newline (\n)
	message, _ = reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	// process for string received, , we should receive the remoteGateway parameter
	m = strings.Split(message, ":")
	if m[0] != "remoteGateway" {
		return fmt.Errorf("Connection error, Received: %s Expected: remoteGateway", m[0])
	}
	s.remoteGateway = strings.TrimSpace(m[1])
	// send string back to client for ACK
	s.conn.Write([]byte(message))
	return nil
}

func (s *Server) createInterface() error {
	// Create TUN interface
	// TODO: Windows have some network specific parameters
	// https://github.com/songgao/water/blob/master/params_windows.go
	var err error
	s.ifce, err = water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return fmt.Errorf("Error creating interface: %v", err)
	}
	log.Printf("Interface Name: %s\n", s.ifce.Name())
	return nil
}

func (s *Server) setupNetwork() error {
	// Create the networking configuration
	dev := defaultInterface
	// Set up routes to remote network depending if we are a server or a client
	s.netCfg = NewNetconfig(s.IfAddress, s.remoteNetwork, s.remoteGateway, s.ifce.Name())
	// The network configuration is deleted when the interface is destroyed
	if err := s.netCfg.SetupNetwork(); err != nil {
		return fmt.Errorf("Error configuting interface network: %v", err)
	}
	log.Printf("Interface Up: %s\n", s.ifce.Name())

	log.Printf("Add route %v\n", s.netCfg.routes)
	if err := s.netCfg.CreateRoutes(); err != nil {
		return fmt.Errorf("Error creating routes: %v", err)
	}
	// Masquerade traffic in server mode and Linux
	log.Printf("Add Masquerade on interface %s\n", dev)
	if err := s.netCfg.CreateMasquerade(dev); err != nil {
		return fmt.Errorf("Error adding masquerade: %v", err)

	}
	return nil
}
