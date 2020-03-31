package main

import (
	"io"
	"log"
	"net"
)

// Tunnel consist in a TCP connection and a HostInterface
// with its networking configuration
type Tunnel struct {
	ifce HostInterface
	conn net.Conn
}

// NewTunnel create a new Tunnel
func NewTunnel(conn net.Conn, ifce HostInterface) *Tunnel {
	log.Println("Creating Tunnel ...")

	return &Tunnel{
		ifce: ifce,
		conn: conn,
	}
}

// Run the Tunnel copies the data from the conn to the interface
// and viceversa
func (t *Tunnel) Run() {
	errCh := make(chan error, 2)
	defer t.ifce.Delete()
	defer t.conn.Close()
	// Copy from the Tun interface to the connection
	go func() {
		for {
			_, err := io.Copy(t.conn, t.ifce.ifce)
			errCh <- err
		}
	}()

	// Copy from the the connection to the Tun interface
	go func() {
		for {
			_, err := io.Copy(t.ifce.ifce, t.conn)
			errCh <- err
		}
	}()

	// Don't fail just log it
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			log.Printf("Tunnel Error: %v", err)
			return
		}
	}
}
