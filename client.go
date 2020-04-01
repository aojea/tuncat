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

// Client represents a client to our server.
type Client struct {
	conn   net.Conn
	ifce   *water.Interface
	netCfg Netconfig
	// Config
	IfAddress     string
	RemoteHost    string
	RemoteNetwork string
	RemoteGateway string
}

// NewClient returns a new instance of Client with default settings.
func NewClient(remoteHost string) *Client {
	return &Client{
		RemoteHost: remoteHost,
	}
}

// Start a new tunnel client
func (c *Client) Start() error {
	var err error
	c.conn, err = net.Dial("tcp", c.RemoteHost)
	if err != nil {
		return fmt.Errorf("Can't connect to server %q: %v", c.RemoteHost, err)
	}
	// Establish the connection: send the tunnel parameters
	errChan := make(chan error, 1)
	timeout := 10 * time.Second
	go func() {
		errChan <- c.handShake()
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
	err = c.createInterface()
	if err != nil {
		log.Fatalf("Error creating Host Interface: %v", err)
	}
	// Configure the interface network
	log.Println("Setup Interface Network...")
	err = c.setupNetwork()
	if err != nil {
		log.Fatalf("Error creating Host Interface: %v", err)
	}
	// Run the tunnel and block
	return Tunnel(c.conn, c.ifce)
}

// Close disconnects the underlying connection to the server.
func (c *Client) Close() {
	log.Println("Shutting down the client...")
	// Close the connection
	if c.conn != nil {
		c.conn.Close()
	}
	// Delete host interface network configuration
	if err := c.netCfg.DeleteRoutes(); err != nil {
		log.Printf("Error deleting routes: %v", err)
	}
	if c.ifce != nil {
		// Close interface
		c.ifce.Close()
	}
}

// handShake do the tunnel connection negotiation sending the configuration parameters for the serveer
// messages has to be echoed from the server in order to the connection to be established
func (c *Client) handShake() error {
	reader := bufio.NewReader(c.conn)
	// Send configuration to the server
	text := fmt.Sprintf("remoteNetwork:%s", c.RemoteNetwork)
	c.conn.Write([]byte(text + "\n"))
	// wait for acknowledge
	message, _ := reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	if strings.TrimSpace(message) != text {
		return fmt.Errorf("Connection error, Sent: %s Received: %s", text, message)
	}
	// Send configuration to the server
	text = fmt.Sprintf("remoteGateway:%s", c.RemoteGateway)
	c.conn.Write([]byte(text + "\n"))
	// wait for acknowledge
	message, _ = reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	if strings.TrimSpace(message) != text {
		return fmt.Errorf("Connection error, Sent: %s Received: %s", text, message)
	}
	return nil
}

func (c *Client) createInterface() error {
	// Create TUN interface
	// TODO: Windows have some network specific parameters
	// https://github.com/songgao/water/blob/master/params_windows.go
	var err error
	c.ifce, err = water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return fmt.Errorf("Error creating interface: %v", err)
	}
	log.Printf("Interface Name: %s\n", c.ifce.Name())
	return nil
}

func (c *Client) setupNetwork() error {
	// Create the networking configuration
	// Set up routes to remote network depending if we are a server or a client
	c.netCfg = NewNetconfig(c.IfAddress, c.RemoteNetwork, c.IfAddress, c.ifce.Name())
	// The network configuration is deleted when the interface is destroyed
	if err := c.netCfg.SetupNetwork(); err != nil {
		return fmt.Errorf("Error configuting interface network: %v", err)
	}
	log.Printf("Interface Up: %s\n", c.ifce.Name())

	log.Printf("Add route %v\n", c.netCfg.routes)
	if err := c.netCfg.CreateRoutes(); err != nil {
		return fmt.Errorf("Error creating routes: %v", err)
	}
	return nil
}
