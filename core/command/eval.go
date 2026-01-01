package command

import (
	"errors"
	"log"
	"net"

	"github.com/vaasu2002/in-memory-storage-engine/core/resp"
)

func evalPING(args []string, c net.Conn) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERROR: Wrong number of arguments for 'ping' command")
	}

	// If no arguemnts respond with PONG
	// else reply with whatever argument was sent
	if len(args) == 0 {
		b = resp.Encode("PONG", true)
	} else {
		b = resp.Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func EvalAndRespond(cmd *KvCmd, c net.Conn) error {
	log.Println("comamnd:", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c) // TODO: Deal with rest of redis commands
	}
}