package lspaxos

import (
	"fmt"
	"net/rpc"
)

type LockOp string
type Err string

const (
	OK             = "OK"
	ErrInvalidLock = "No such lock to unlock"
	ErrLockHeld    = "Lock held by someone else"
)

const (
	Unlock LockOp = "Unlock"
	Lock   LockOp = "Lock"
)

type Command struct {
	// Lock name
	LockName string

	// Lock operation
	LockOp LockOp

	// Message Id
	MsgId int

	// Client Id
	ClientId int64
}

// Number.Leader, Number takes precedence
type Ballot struct {
	Number int
	Leader int
}

// Client request response from the replica
type ClientRequest struct {
	// Command
	Command Command
}

type ClientResponse struct {
	// Error string
	Err Err
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
}

func Call(
	ServerAddress string,
	ProcedureName string,
	Request interface{},
	Response interface{},
	Done chan interface{},
) {
	var client, err = rpc.Dial("tcp", ServerAddress)
	defer client.Close()
	if err != nil {
		fmt.Printf("Error on Dial() Server:%s Procedure:%s\n", ServerAddress, ProcedureName)
		return
	}
	err = client.Call(ProcedureName, Request, Response)
	if err != nil {
		fmt.Printf("Error on Dial() Server:%s Procedure:%s Error: %s\n", ServerAddress, ProcedureName, err)
		return
	}

	Done <- Response
}

type TestRPCHandler struct{}

func (h *TestRPCHandler) Execute(req Ballot, res *Ballot) (err error) {
	//if req.Number != 69 {
	//	err = errors.New("Wrong key!!")
	//	return
	//}
	fmt.Printf("In Execute, %+v\n", req)
	res.Leader = req.Number
	return
}
