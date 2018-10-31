package main

import (
	"fmt"
	"lspaxos"
)

func main() {
	TestRpcClient("1234")
}

func TestRpcClient(ServerPort string) {
	var addr = "172.28.1.105:" + ServerPort	
	ch := make(chan interface{}, 5)
	requestNumbers := [5]int{1, 2, 3, 4, 5}
	
	for i := 0; i < 5; i++ {
	    request := &lspaxos.Ballot{Number: requestNumbers[i]}
	    response := new(lspaxos.Ballot)
            go lspaxos.Call(addr, "TestRPCHandler.Execute", request, response, ch)
	}
	
	for i := 0; i < 5; i++ {
	    resp := <- ch
	    if resp == false {
	       fmt.Printf("Response %d failed...\n", i)
	       continue
	    }

	    fmt.Printf("Response %d: %+v\n", i, resp)
	}
}
