package command

import (
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/vaasu2002/in-memory-storage-engine/core"
	"github.com/vaasu2002/in-memory-storage-engine/core/resp"
)

var RESP_NIL []byte = []byte("$-1\r\n")


func evalPING(args []string, c io.ReadWriter) error {
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

func evalSET(args []string, c io.ReadWriter) error {

	if len(args) < 2 {
		return errors.New("Error: Less number of arguments for 'SET' command")
	}

	var key, value string
	var exDurationMs int64 = -1 // Default: No expiry
	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		
		// Expliry command
		case "EX", "ex" :
			i++
			if i == len(args) {
				return errors.New("Error: Less number of arguments for `expiry` command")
			}
			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return errors.New("Error: Value is not an integer or out of range")
			}
			exDurationMs = exDurationSec * 1000

		default:
			return errors.New("Error: Invalid arguments. Sytax error")
		}
	}
	
	core.Put(key, core.NewObj(value, exDurationMs))
	c.Write([]byte("+OK\r\n"))
	return nil
}

func EvalAndRespond(cmd *KvCmd, c io.ReadWriter) error {
	log.Println("comamnd:", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c) // TODO: Deal with rest of redis commands
	}
}