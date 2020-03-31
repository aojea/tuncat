package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// ClientConnect do the tunnel connection negotiation sending the configuration parameters for the serveer
// messages has to be echoed from the server in order to the connection to be established
func ClientConnect(conn net.Conn, remoteNetwork, remoteGateway string) error {
	reader := bufio.NewReader(conn)
	// Send configuration to the server
	text := fmt.Sprintf("remoteNetwork:%s", remoteNetwork)
	conn.Write([]byte(text + "\n"))
	// wait for acknowledge
	message, _ := reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	if strings.TrimSpace(message) != text {
		return fmt.Errorf("Connection error, Sent: %s Received: %s", text, message)
	}
	// Send configuration to the server
	text = fmt.Sprintf("remoteGateway:%s", remoteGateway)
	conn.Write([]byte(text + "\n"))
	// wait for acknowledge
	message, _ = reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	if strings.TrimSpace(message) != text {
		return fmt.Errorf("Connection error, Sent: %s Received: %s", text, message)
	}
	return nil
}

// ServerConnect do the tunnel connection negotiation
func ServerConnect(conn net.Conn) (string, string, error) {
	reader := bufio.NewReader(conn)
	// Receive the configuration parameters
	// will listen for message to process ending in newline (\n)
	message, _ := reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	// process for string received, we should receive the remoteNetwork parameter
	m := strings.Split(message, ":")
	if m[0] != "remoteNetwork" {
		return "", "", fmt.Errorf("Connection error, Received: %s Expected: remoteNetwork", m[0])
	}
	remoteNetwork := strings.TrimSpace(m[1])
	// send string back to client for ACK
	conn.Write([]byte(message))
	// will listen for message to process ending in newline (\n)
	message, _ = reader.ReadString('\n')
	// output message received
	log.Printf("Message Received: %s", message)
	// process for string received, , we should receive the remoteGateway parameter
	m = strings.Split(message, ":")
	if m[0] != "remoteGateway" {
		return "", "", fmt.Errorf("Connection error, Received: %s Expected: remoteGateway", m[0])
	}
	remoteGateway := strings.TrimSpace(m[1])
	// send string back to client for ACK
	conn.Write([]byte(message))

	return remoteNetwork, remoteGateway, nil
}
