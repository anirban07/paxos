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
	var addr = "localhost:" + ServerPort

	done := make(chan *rpc.Call, 10)
	reqVals := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	reqValToRes := make(map[lspaxos.Ballot]*lspaxos.Ballot)

	for _, reqVal := range reqVals {
		req := lspaxos.Ballot{Number: reqVal, Leader: 0}
		res := new(lspaxos.Ballot)
		reqValToRes[req] = res
	}

	for req, res := range reqValToRes {
		go lspaxos.Call(addr, "TestRPCHandler.Execute", req, res, done)
	}

	// time.Sleep(6 * time.Second)
	for range reqVals {
		resCall := <-done
		// req, _ := resCall.Args.(lspaxos.Ballot)
		// res := reqValToRes[req.Number]
		fmt.Println("Got reply:", resCall.Reply, "error: ", resCall.Error)
	}
}
