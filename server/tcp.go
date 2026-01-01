package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"github.com/vaasu2002/in-memory-storage-engine/config"
)

// TODO: This max read in one shot is 512 Bytes, to allow input > 512 Bytes, repeat read until EOF.
func readCommand(c net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}


func respond(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

func RunSyncTcpServer() {
	log.Println("Starting a synchronous TCP server on", config.Host, config.Port)

	var connClients int = 0

	// listening
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	// Infinite loop listening to the client connections
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		connClients += 1
		log.Println("Client connected with address:", c.RemoteAddr(), " | Concurrent clients: ", connClients)

		// Infinite loop continuously reading the command over the socket
		for {
			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				connClients -= 1
				log.Println("Client disconnected:", c.RemoteAddr(), " | Concurrent clients: ", connClients)
				if err == io.EOF {
					break
				}
				log.Println("Error: ", err)
			}
			log.Println("Command: ", cmd)
			
			// Responding with same command to client (echoing)
			if err = respond(cmd, c); err != nil {
				log.Print("err write:", err)
			}
		}
	}
}