package main

import (
	"flag" // pkg for parsing command-line arguments
	"log"

	"github.com/vaasu2002/in-memory-storage-engine/config"
	"github.com/vaasu2002/in-memory-storage-engine/server"
)

func setupFlags() {
	// A host is an IP address that identifies where your server should listen for connections.
	// (0.0.0.0) means it listens from all the IPs
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for in-memory server")
	flag.IntVar(&config.Port, "port", 8379, "Port for the server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("Starting the storage engine....")
	server.RunAsyncTcpServer()
}