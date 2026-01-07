package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"syscall"

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
func readCommand(c io.ReadWriter) (*core.KvCmd, error) {
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

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd *core.KvCmd, c io.ReadWriter) {
	err := core.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
}

var connClients int = 0
const maxClients int  = 20000

func RunAsyncTcpServer() error {
	
	log.Println("Starting a asynchronous TCP server on", config.Host, config.Port)

	// Create EPOLL event object (buffer) to hold events (file descriptors)
	// that are ready for I/O.
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, maxClients)

	// Create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK | syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	// Ensure the server socket is closed when the function exits
	defer syscall.Close(serverFD)

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP address and port to the socket's file descriptor.
	ipv4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	}); err != nil {
		return err
	}

	// The socket will listen on the specified IP address and port.
	if err = syscall.Listen(serverFD, maxClients); err != nil {
		return err
	}

	// Creating Epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// Registering the server socket with the Epoll instance
	// to monitor incoming connection requests.
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd: int32(serverFD),
	}

	// Listen to read events on the server (socket)
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	// Event loop (infinite loop)
	for {
		// Blocking call: Wait until any client/s (represented by file descriptor) is/area
		// ready for I/O. Thread sleep until they are available.
		nevents, e := syscall.EpollWait(epollFD, events[:], -1)
		if e != nil {
			continue
		}

		for i:=0; i < nevents; i++ {

			// To check if server got a new connection request from a client
			// check if the event is for the server socket.
			// Accept the connection and add the client socket to epoll 
			// for monitoring
			if int(events[i].Fd) == serverFD {
				// accept the incoming connection from a client
				clientFd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Panicln("Error: ", err)
					continue
				}
				connClients++
				syscall.SetNonblock(clientFd, true)

				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(clientFd),
				}

				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, clientFd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}

			} else {
				
				comm := core.FDComm {
					Fd : int(events[i].Fd),
				}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					connClients--;
					continue;
				}
				respond(cmd, comm)
			}
		}
	}
}