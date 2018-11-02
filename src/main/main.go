package main

import (
	"fmt"
	"lspaxos"
	"net"
	"net/rpc"
	"os"
)

func main() {
	TestRpcServer(os.Args[1])
}

func TestRpcServer(Port string) {
	rpc.Register(&lspaxos.TestRPCHandler{})
	fmt.Println("Listening on " + Port)
	listener, _ := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	rpc.Accept(listener)
}
