package lspaxos

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
)

// Defines an the Acceptor state
// - Keeps track of a ballot number (highest seen)
// - Keeps track of a map of previously accepted commands (if any)
type AcceptorServer struct {
	// Highest ballot number promised by this acceptor
	Ballot Ballot

	// Map to store slot number with commands
	AcceptedValues map[int]Command
}

// Handler for Scout RPC's
// Checks if ballot number is higher and updates if necessary
func (h *AcceptorServer) ExecutePropose(req ScoutRequest, res *ScoutResponse) (err error) {
	if CompareBallot(req.Ballot, h.Ballot) > 0 {
		h.Ballot = req.Ballot
		// For debug purposes, TODO: Anir, is there a better way to log this?
		fmt.Printf("Updated ballot %+v, accepted: %+v\n", h.Ballot, h.AcceptedValues)
	}

	res.Ballot = h.Ballot
	res.AcceptedValues = h.AcceptedValues
	return
}

// Handler for Commander RPC's
// Only accepts the command sent by the commander if the ballot of the commander is equivalent
// to the ballot promised by the acceptor (i.e. The Commander is the leader)
func (h *AcceptorServer) ExecuteAccept(req CommanderRequest, res *CommanderResponse) (err error) {
	if CompareBallot(req.Ballot, h.Ballot) == 0 {
		h.AcceptedValues[req.Slot] = req.Command
	}

	res.Ballot = h.Ballot
	return
}

// Helper to spawn an instance of an Acceptor
// Example usage: go StartAcceptorServer("Alpha", 1234)
// Returns an error if unable to set up a listening port
func StartAcceptorServer(Name string, Port string) (err error) {
	rpc.Register(&AcceptorServer{AcceptedValues: make(map[int]Command)})
	fmt.Printf("Acceptor server %s listening on port %s\n", Name, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on " + Name)
		return
	}

	rpc.Accept(listener)
	return
}
