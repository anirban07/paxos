package lspaxos

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
)

const (
	ReplicaResponsesChannelSize = 512
)

type Replica struct {
	// UNGAURDED ACCES PERMITTER TO THIS FIELD
	// Channel ReplicaResponse from Leader
	replicaResponses chan interface{}

	// Mutex to synchronize access to the fields below
	mu sync.Mutex

	// Unique identifier of the replica
	replicaID int64

	// Map from lock name to client holding it
	lockMap map[string]int64

	// The index of the next slot in which the replica
	// has not yet proposed any command, initially 1
	slotIn int

	// The index of the next slot for which it needs to learn a
	// decision before it can update its copy of the application state
	slotOut int

	// Condition variable for when a new item is added to requests
	newRequest sync.Cond

	// Client requests not yet proposed or decided
	requests []Command

	// Proposals that are currently outstanding
	proposals map[int]Command

	// Condition variable for when command is performed
	somethingPerformed sync.Cond

	// Proposals that are known to have been decided
	decisions map[int]Command

	// Leaders
	leaders []string
}

func (thisReplica *Replica) propose() {
	for {
		thisReplica.mu.Lock()
		for len(thisReplica.requests) == 0 {
			thisReplica.newRequest.Wait()
		}
		// Propose values. Should empty out the Requests
		for _, requestCommand := range thisReplica.requests {
			request := ReplicaRequest{
				Command: requestCommand,
				Slot:    thisReplica.slotIn,
			}
			for _, leader := range thisReplica.leaders {
				response := new(ReplicaResponse)
				go Call(
					leader,
					"Leader.ExecutePropose",
					request,
					response,
					thisReplica.replicaResponses,
				)
			}
			thisReplica.proposals[thisReplica.slotIn] = requestCommand
			thisReplica.slotIn++
		}
		thisReplica.requests = nil
		thisReplica.mu.Unlock()
	}
}

func (thisReplica *Replica) perform() {
	for {
		response := <-thisReplica.replicaResponses
		if response == false {
			// TODO: Do we expect a failure here?
			continue
		}
		replicaResponse := response.(ReplicaResponse)
		thisReplica.mu.Lock()
		thisReplica.decisions[replicaResponse.Slot] = replicaResponse.Command
		decidedCommand, present := thisReplica.decisions[thisReplica.slotOut]
		for present {
			thisReplica.somethingPerformed.Broadcast()
			proposedCommand := thisReplica.proposals[thisReplica.slotOut]
			delete(thisReplica.proposals, thisReplica.slotOut)
			if !decidedCommand.Equals(proposedCommand) {
				thisReplica.requests = append(thisReplica.requests, proposedCommand)
				thisReplica.newRequest.Signal()
			}
			lockOwner, lockIsOwned := thisReplica.lockMap[decidedCommand.LockName]
			switch decidedCommand.LockOp {
			case Lock:
				if !lockIsOwned || lockOwner == decidedCommand.ClientID {
					thisReplica.lockMap[decidedCommand.LockName] = decidedCommand.ClientID
				}
			case Unlock:
				if lockIsOwned && lockOwner == decidedCommand.ClientID {
					delete(thisReplica.lockMap, decidedCommand.LockName)
				}
			}
			thisReplica.slotOut++
			decidedCommand, present = thisReplica.decisions[thisReplica.slotOut]
		}
		thisReplica.mu.Unlock()
	}
}

func StartReplica(ReplicaID int64, Leaders []string, Port string) (err error) {
	thisReplica := &Replica{
		mu:               sync.Mutex{},
		replicaResponses: make(chan interface{}, ReplicaResponsesChannelSize),
		replicaID:        ReplicaID,
		lockMap:          make(map[string]int64),
		slotIn:           1,
		slotOut:          1,
		requests:         make([]Command, 8),
		proposals:        make(map[int]Command),
		decisions:        make(map[int]Command),
		leaders:          Leaders,
	}
	thisReplica.newRequest = sync.Cond{L: &thisReplica.mu}
	thisReplica.somethingPerformed = sync.Cond{L: &thisReplica.mu}
	rpc.Register(thisReplica)
	log.Printf("Replica %d listening on port %s\n", ReplicaID, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()
	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on replica " + string(ReplicaID))
		return err
	}

	go thisReplica.propose()
	rpc.Accept(listener)
	return nil
}

func (thisReplica *Replica) ExecuteRequest(req ClientRequest, res *ClientResponse) (err error) {
	// TODO: Check for duplicate req in Requests, Proposals, Decisions
	res.MsgID = req.Command.MsgID
	requestDecided := false
	thisReplica.mu.Lock()
	thisReplica.requests = append(thisReplica.requests, req.Command)
	thisReplica.mu.Unlock()
	thisReplica.newRequest.Signal()

	for !requestDecided {
		thisReplica.mu.Lock()
		thisReplica.somethingPerformed.Wait()
		for slot, decidedCommand := range thisReplica.decisions {
			if req.Command.Equals(decidedCommand) && slot < thisReplica.slotOut {
				lockOwner, lockIsOwned := thisReplica.lockMap[req.Command.LockName]
				switch req.Command.LockOp {
				case Lock:
					if lockIsOwned && lockOwner == req.Command.ClientID {
						res.Err = OK
					} else {
						res.Err = ErrLockHeld
					}
				case Unlock:
					if lockIsOwned && lockOwner != req.Command.ClientID {
						res.Err = ErrInvalidUnlock
					} else {
						res.Err = OK
					}
				}
				requestDecided = true
				break
			}
		}
		thisReplica.mu.Unlock()
	}
	return nil
}
