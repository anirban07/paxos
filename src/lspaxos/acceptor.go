package lspaxos

import (
	"errors"
	"log"
	"net"
	"net/rpc"
)

// Defines an the Acceptor state
// - Keeps track of a ballot number (highest seen)
// - Keeps track of a map of previously accepted commands (if any)
type Acceptor struct {
	// Unique identifier of the acceptor
	AcceptorID int64

	// Highest ballot number promised by this acceptor
	Ballot Ballot

	// Map to store slot number with commands
	AcceptedValues map[int]Command
}

// Handler for Scout RPC's
// Checks if ballot number is higher and updates if necessary
func (thisAcceptor *Acceptor) ExecutePropose(req ScoutRequest, res *ScoutResponse) (err error) {
	if req.Ballot.Compare(thisAcceptor.Ballot) > 0 {
		thisAcceptor.Ballot = req.Ballot
		// For debug purposes, TODO: Anir, is there a better way to log this?
		log.Printf(
			"Acceptor %d updated ballot %+v, accepted: %+v\n",
			thisAcceptor.AcceptorID,
			thisAcceptor.Ballot,
			thisAcceptor.AcceptedValues,
		)
	}

	res.Ballot = thisAcceptor.Ballot
	res.AcceptedValues = thisAcceptor.AcceptedValues
	return nil
}

// Handler for Commander RPC's
// Only accepts the command sent by the commander if the ballot of the commander is equivalent
// to the ballot promised by the acceptor (i.e. The Commander is the leader)
func (thisAcceptor *Acceptor) ExecuteAccept(req CommanderRequest, res *CommanderResponse) (err error) {
	if req.Ballot.Compare(thisAcceptor.Ballot) == 0 {
		thisAcceptor.AcceptedValues[req.Slot] = req.Command
		log.Printf(
			"Acceptor %d accepted ballot:%+v slot:%d command:%+v \n",
			thisAcceptor.AcceptorID,
			req.Ballot,
			req.Slot,
			req.Command,
		)
	}

	res.Ballot = thisAcceptor.Ballot
	return nil
}

// Helper to spawn an instance of an Acceptor
// Example usage: go StartAcceptorServer("Alpha", 1234)
// Returns an error if unable to set up a listening port
func StartAcceptor(AcceptorID int64, Port string) (err error) {
	rpc.Register(&Acceptor{
		AcceptorID:     AcceptorID,
		Ballot:         Ballot{-1, -1},
		AcceptedValues: make(map[int]Command),
	})
	log.Printf("Acceptor %d listening on port %s\n", AcceptorID, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on " + string(AcceptorID))
		return err
	}

	rpc.Accept(listener)
	return nil
}
