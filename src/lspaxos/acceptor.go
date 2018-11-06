package lspaxos

import (
	"log"
	"net"
	"net/rpc"
	"sync/atomic"
)

// Defines an the Acceptor state
// - Keeps track of a ballot number (highest seen)
// - Keeps track of a map of previously accepted commands (if any)
type Acceptor struct {
	// Unique identifier of the acceptor
	acceptorID int

	// Highest ballot number promised by this acceptor
	ballot Ballot

	// Map to store slot number with commands
	acceptedValues map[int]Command

	// Listener
	listener net.Listener

	// For debugging
	dead int32
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

func (thisAcceptor *Acceptor) kill() {
	log.Printf("Killing acceptor %d\n", thisAcceptor.acceptorID)
	atomic.StoreInt32(&thisAcceptor.dead, 1)
	if thisAcceptor.listener != nil {
		thisAcceptor.listener.Close()
	}
}

func (thisAcceptor *Acceptor) isDead() bool {
	return atomic.LoadInt32(&thisAcceptor.dead) != 0
}

//StartAcceptor starts an acceptor instance and returns an Acceptor struct.
//The struct can be used to kill this instance.
func StartAcceptor(AcceptorID int, Port string) (acceptor *Acceptor) {
	server := rpc.NewServer()
	listener, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf(
			"Acceptor %d failed to set up listening port %s\n",
			AcceptorID,
			Port,
		)
		return nil
	}
	acceptor = &Acceptor{
		acceptorID:     AcceptorID,
		ballot:         Ballot{-1, -1},
		acceptedValues: make(map[int]Command),
		listener:       listener,
		dead:           0,
	}
	server.Register(acceptor)

	go func() {
		for !acceptor.isDead() {
			log.Printf("Acceptor %d listening for requests\n", AcceptorID)
			connection, err := acceptor.listener.Accept()
			if err == nil {
				log.Printf("Acceptor accepted request\n")
				server.ServeConn(connection)
			} else {
				log.Fatalf("Acceptor %d failed to accept connection\n", AcceptorID)
			}
		}
	}()
	return acceptor
}
