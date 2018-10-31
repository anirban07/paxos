package lspaxos

import (
	"net"
	"net/rpc"
	"testing"
)

type TestRPCHandler struct {
}

func TestRpc(t *testing.T) {
	rpc.Register(&TestRPCHandler{})
	Port := "1234"
	t.Log("Listening on " + Port)
	listener, _ := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	rpc.Accept(listener)
}
