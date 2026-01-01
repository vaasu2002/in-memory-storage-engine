package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/vaasu2002/in-memory-storage-engine/config"
	"github.com/vaasu2002/in-memory-storage-engine/core"
)

// TODO: This max read in one shot is 512 Bytes, to allow input > 512 Bytes, repeat read until EOF.
// Reading all the bytes that are coming in from the client into a buffer
// then decoding it into an array string
// When cli or any clinet wants to issue command to this server
// command typically has root command and arguments like PUT a,10
// This is sent to the server as array of bytes
// we convert it to array of strings
func readCommand(c net.Conn) (*core.KvCmd, error) {
	// Make buffer of 512 bytes
	var buf []byte = make([]byte, 512)
	// Read from the socket and store in buffer
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}
	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		log.Println("decode error:", err)
		return nil, err
	}
	log.Printf("decoded tokens: %#v", tokens)

	return &core.KvCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respondError(err error, c net.Conn) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.KvCmd, c net.Conn) {
	err := core.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
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
			respond(cmd, c)
		}
	}
}
