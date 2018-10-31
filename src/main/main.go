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
	TestRpcServer(os.Args[1])
}

func TestRpcServer(Port string) {
	rpc.Register(&lspaxos.TestRPCHandler{})
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
