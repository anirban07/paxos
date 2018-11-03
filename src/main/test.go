package main

import (
	"fmt"
	"lspaxos"
)

func main() {
	TestAcceptor("1234")
}

func TestAcceptor(ServerPort string) {
	var addr = "localhost:" + ServerPort
	ch := make(chan interface{}, 5)
	requestNumbers := [5]int{1, 1, 1, 1, 1}

	// Send contending ballots from scouts 1-5
	for i := 0; i < 5; i++ {
		ballot := lspaxos.Ballot{Number: requestNumbers[i], Leader: i + 1}
		request := &lspaxos.ScoutRequest{Ballot: ballot}
		response := new(lspaxos.ScoutResponse)
		go lspaxos.Call(addr, "Acceptor.ExecutePropose", request, response, ch)
	}

	for i := 0; i < 5; i++ {
		resp := <-ch
		if resp == false {
			fmt.Printf("Response %d failed...\n", i)
			continue
		}

		// The map should be empty for all responses
		fmt.Printf("Response %d: %+v\n", i, resp)
	}

	// Send commander request
	ballot := lspaxos.Ballot{Number: 1, Leader: 5}
	command := lspaxos.Command{LockName: "Theta", LockOp: lspaxos.Unlock, MsgID: 1, ClientID: 5}
	request := &lspaxos.CommanderRequest{Ballot: ballot, Slot: 1, Command: command}
	response := new(lspaxos.CommanderResponse)
	go lspaxos.Call(addr, "Acceptor.ExecuteAccept", request, response, ch)

	resp := <-ch
	fmt.Printf("Response: %+v\n", resp)
}
