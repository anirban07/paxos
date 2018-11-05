package lspaxos

import (
	"log"
	"net/rpc"
)

type LockOp string
type Err string

const (
	OK               = "OK"
	ErrInvalidUnlock = "Attempting to unlock unheld lock"
	ErrLockHeld      = "Lock held by someone else"
)

const (
	Unlock            LockOp = "Unlock"
	Lock              LockOp = "Lock"
	ChannelBufferSize        = 512
)

type Command struct {
	// Lock name
	LockName string

	// Lock operation
	LockOp LockOp

	// Message Id
	MsgID int

	// Client Id
	ClientID int64
}

func (this Command) Equals(other Command) bool {
	return this.ClientID == other.ClientID && this.MsgID == other.MsgID
}

// Number.Leader, Number takes precedence
type Ballot struct {
	Number int
	Leader int
}

// Compares two ballot numbers
func (this Ballot) Compare(other Ballot) int {
	if this.Number == other.Number {
		return this.Leader - other.Leader
	}
	return this.Number - other.Number
}

// Client request response from the replica
type ClientRequest struct {
	// Command
	Command Command
}

type ClientResponse struct {
	// Error string
	Err Err

	// Used to verify on the client side
	MsgID int
}

// Replica-Leader request/response

// This is sent to the leader from the replica when the replica
// gets a new request that it has not seen before from a client.
type ReplicaRequest struct {
	// Command to be proposed
	Command Command

	// Slot number (slot to be proposed)
	Slot int
}

// This is sent to the replica from the Commander after a slot
// has been decided. (Commander must send this)
type ReplicaResponse struct {
	// Command that was decided
	Command Command

	// Slot number (slot that was decided)
	Slot int
}

// Commander to Acceptor

// The Commander sends this to all acceptors when the Commander
// has spawned.
type CommanderRequest struct {
	// Command that is being proposed to the acceptors
	Command Command

	// Slot number (slot that is being proposed)
	Slot int

	// Ballot number
	Ballot Ballot
}

// Response received by the Commander from the Acceptor
type CommanderResponse struct {
	Ballot Ballot

	AcceptorID int64
}

// Scout to Acceptor

// Scout sends a ballot number to all acceptors
type ScoutRequest struct {
	Ballot Ballot
}

// Scout gets back the highest accepted ballot number by an acceptor
// as well as all previous values accepted by the acceptor
type ScoutResponse struct {
	Ballot Ballot

	AcceptedValues map[int]Command

	AcceptorID int64
}

// Call is a wrapper function for creating a connection to a remote
// server and making an RPC.
// It is a blocking operation, and the caller should use a goroutine
// to call it.
// It sends the response in the Done channel on success, or false on
// failure
func Call(
	ServerAddress string,
	ProcedureName string,
	Request interface{},
	Response interface{},
	Done chan interface{},
) {
	client, err := rpc.Dial("tcp", ServerAddress)
	defer client.Close()
	if err != nil {
		log.Printf(
			"Error on Dial() Server:%s Procedure:%s\n",
			ServerAddress,
			ProcedureName,
		)
		Done <- false
		return
	}
	err = client.Call(ProcedureName, Request, Response)
	if err != nil {
		log.Printf(
			"Error on Call() Server:%s Procedure:%s Error: %s\n",
			ServerAddress,
			ProcedureName,
			err,
		)
		Done <- false
		return
	}

	Done <- Response
}
