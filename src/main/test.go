package main

import (
	"fmt"
	"lspaxos"
)

func main() {
	TestRpcClient("1234")
}

func TestRpcClient(ServerPort string) {
	var addr = "klassert.cs.washington.edu:" + ServerPort
	var request = &lspaxos.Ballot{Number: 69}
	var response = new(lspaxos.Ballot)

	lspaxos.Call(addr, "TestRPCHandler.Execute", request, response)
	fmt.Println("~~~")
	fmt.Println(response)
}
