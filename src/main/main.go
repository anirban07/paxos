package main

import (
	"fmt"
	"lspaxos"
	"net"
	"net/rpc"
	"os"
)

func main() {
        lspaxos.StartAcceptorServer("Alpha", os.Args[1])
	//TestRpcServer(os.Args[1])
}

func TestRpcServer(Port string) {
	rpc.Register(&lspaxos.AcceptorServer{})
	fmt.Println("Listening on " + Port)
	listener, _ := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	rpc.Accept(listener)
}
