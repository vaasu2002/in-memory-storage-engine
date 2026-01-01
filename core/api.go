package core

import (
	"net"

	"github.com/vaasu2002/in-memory-storage-engine/core/command"
	"github.com/vaasu2002/in-memory-storage-engine/core/resp"
)

type KvCmd = command.KvCmd

func DecodeArrayString(b []byte) ([]string, error) {
	return resp.DecodeArrayString(b)
}

func EvalAndRespond(cmd *KvCmd, c net.Conn) error {
	return command.EvalAndRespond(cmd, c)
}