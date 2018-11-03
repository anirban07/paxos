package main

import (
	"fmt"
	"lspaxos"
	"net"
	"net/rpc"
	"os"
)

// type TestRPCHandler struct {
// }

// type TestRequest struct {
// 	Name string
// }

// type TestResponse struct {
// 	Message string
// }

func main() {
    lspaxos.StartAcceptorServer("Alpha", os.Args[1])
	//TestRpcServer(os.Args[1])
	/*var commands = lspaxos.ReadSpec("../specs/test_spec.txt")
	replicas := [1]string {"Delta"}
	for _, command := range commands {
		fmt.Println(command)
	}

	var client = lspaxos.Client{ClientId: 2048, Replicas: replicas[:]}
	client.StartClient(commands)*/
}

func TestRpcServer(Port string) {
	rpc.Register(&lspaxos.AcceptorServer{})
	fmt.Println("Listening on " + Port)
	listener, _ := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	rpc.Accept(listener)
}

// func (h *TestRPCHandler) Execute(req TestRequest, res *TestResponse) (err error) {
// 	if req.Name != "SecretKey" {
// 		err = errors.New("Wrong key!!")
// 		return
// 	}
// 	res.Message = "SecretValue"
// 	return
// }
