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
	acceptorID int64

	// Highest ballot number promised by this acceptor
	ballot Ballot

	// Map to store slot number with commands
	acceptedValues map[int]Command
}

// Handler for Scout RPC's
// Checks if ballot number is higher and updates if necessary
func (thisAcceptor *Acceptor) ExecutePropose(req ScoutRequest, res *ScoutResponse) (err error) {
	log.Printf("Acceptor %d got a propose request %+v\n", thisAcceptor.acceptorID, req)
	if req.Ballot.Compare(thisAcceptor.ballot) > 0 {
		thisAcceptor.ballot = req.Ballot
		log.Printf(
			"Acceptor %d updated ballot %+v, accepted: %+v\n",
			thisAcceptor.acceptorID,
			thisAcceptor.ballot,
			thisAcceptor.acceptedValues,
		)
	}

	res.Ballot = thisAcceptor.ballot
	res.AcceptedValues = thisAcceptor.acceptedValues
	res.AcceptorID = thisAcceptor.acceptorID
	return nil
}

// Handler for Commander RPC's
// Only accepts the command sent by the commander if the ballot of the commander is equivalent
// to the ballot promised by the acceptor (i.e. The Commander is the leader)
func (thisAcceptor *Acceptor) ExecuteAccept(req CommanderRequest, res *CommanderResponse) (err error) {
	log.Printf("Acceptor %d got an accept request %+v\n", thisAcceptor.acceptorID, req)
	if req.Ballot.Compare(thisAcceptor.ballot) >= 0 {
		thisAcceptor.ballot = req.Ballot
		thisAcceptor.acceptedValues[req.Slot] = req.Command
		log.Printf(
			"Acceptor %d accepted ballot:%+v slot:%d command:%+v \n",
			thisAcceptor.acceptorID,
			req.Ballot,
			req.Slot,
			req.Command,
		)
	}

	res.Ballot = thisAcceptor.ballot
	res.AcceptorID = thisAcceptor.acceptorID
	return nil
}

// Helper to spawn an instance of an Acceptor
// Example usage: go StartAcceptorServer("Alpha", 1234)
// Returns an error if unable to set up a listening port
func StartAcceptor(AcceptorID int64, Port string) (err error) {
	server := rpc.NewServer()
	server.Register(&Acceptor{
		acceptorID:     AcceptorID,
		ballot:         Ballot{-1, -1},
		acceptedValues: make(map[int]Command),
	})
	log.Printf("Acceptor %d listening on port %s\n", AcceptorID, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()

	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on acceptor " + string(AcceptorID))
		return err
	}

	server.Accept(listener)
	return nil
}
