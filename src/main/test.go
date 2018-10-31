package main

import (
	"fmt"
	"lspaxos"
	"net/rpc"
)

func main() {
	TestRpcClient("1234")
}

func TestRpcClient(ServerPort string) {
	var addr = "klassert.cs.washington.edu:" + ServerPort
	var request = &lspaxos.Ballot{Number: 69}
	var response = new(lspaxos.Ballot)

	var client, _ = rpc.Dial("tcp", addr)
	defer client.Close()
	err := client.Call("TestRPCHandler.Execute", request, response)
	fmt.Println(err)
	fmt.Println("~~~")
	fmt.Println(response)
}
