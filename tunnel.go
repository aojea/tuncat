package main

import (
	"fmt"
	"io"
	"net"

	"github.com/songgao/water"
)

// Tunnel copies the data from the conn to the interface
// and viceversa
func Tunnel(conn net.Conn, ifce *water.Interface) error {
	errCh := make(chan error, 2)
	// Copy from the Tun interface to the connection
	go func() {
		for {
			_, err := io.Copy(conn, ifce)
			errCh <- err
		}
	}()

	// Copy from the the connection to the Tun interface
	go func() {
		for {
			_, err := io.Copy(ifce, conn)
			errCh <- err
		}
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			return fmt.Errorf("Tunnel Error: %v", err)
		}
	}
	return nil
}
