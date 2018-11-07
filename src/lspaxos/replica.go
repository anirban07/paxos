package lspaxos

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
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
	replicaID int

	// Map from lock name to client holding it
	lockMap map[string]int

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

	// Listener
	listener net.Listener

	// Address
	Address string

	// For debugging
	dead int32
}

func (thisReplica *Replica) propose() {
	for {
		thisReplica.mu.Lock()
		log.Printf("Replica %d has proposals %+v\n", thisReplica.replicaID, thisReplica.proposals)
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
			continue
		}
		replicaResponse := response.(*ReplicaResponse)
		log.Printf("Replica %d got a response %+v\n", thisReplica.replicaID, replicaResponse)
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

func (thisReplica *Replica) ExecuteRequest(req ClientRequest, res *ClientResponse) (err error) {
	// TODO: Check for duplicate req in Requests, Proposals, Decisions
	// This is impossible
	log.Printf("Replica %d got a request %+v\n", thisReplica.replicaID, req.Command)
	res.MsgID = req.Command.MsgID
	requestDecided := false
	thisReplica.mu.Lock()
	thisReplica.requests = append(thisReplica.requests, req.Command)
	thisReplica.mu.Unlock()
	thisReplica.newRequest.Signal()

	for !requestDecided {
		thisReplica.mu.Lock()
		thisReplica.somethingPerformed.Wait()
		localLockMap := make(map[string]int)
		var logState = fmt.Sprintf("Replica %d state:", thisReplica.replicaID)
		for slot := 0; slot < thisReplica.slotOut; slot++ {
			decidedCommand := thisReplica.decisions[slot]
			var result Err = ErrConnectionError
			lockOwner, lockIsOwned := localLockMap[decidedCommand.LockName]
			switch decidedCommand.LockOp {
			case Lock:
				if !lockIsOwned || lockOwner == decidedCommand.ClientID {
					localLockMap[decidedCommand.LockName] = decidedCommand.ClientID
					result = OK
					logState += fmt.Sprintf(" (%d %+v)", slot, decidedCommand)
				} else {
					result = ErrLockHeld
				}
			case Unlock:
				if !lockIsOwned || lockOwner != decidedCommand.ClientID {
					result = ErrInvalidUnlock
				} else {
					delete(localLockMap, decidedCommand.LockName)
					result = OK
					logState += fmt.Sprintf(" (%d %+v)", slot, decidedCommand)
				}
			}
			if decidedCommand.Equals(req.Command) {
				requestDecided = true
				res.Err = result
				break
			}
		}
		log.Printf(logState)
		thisReplica.mu.Unlock()
	}
	return nil
}

func (thisReplica *Replica) kill() {
	log.Printf("Killing replica %d\n", thisReplica.replicaID)
	atomic.StoreInt32(&thisReplica.dead, 1)
	if thisReplica.listener != nil {
		thisReplica.listener.Close()
	}
}

func (thisReplica *Replica) isDead() bool {
	return atomic.LoadInt32(&thisReplica.dead) != 0
}

//StartReplica starts an acceptor instance and returns an Replica struct.
//The struct can be used to kill this instance.
func StartReplica(ReplicaID int, LeaderAddresses []string, Address string) (replica *Replica) {
	server := rpc.NewServer()
	listener, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalf(
			"Replica %d failed to set up listening address %s, %s\n",
			ReplicaID,
			Address,
			err,
		)
		return nil
	}
	replica = &Replica{
		mu:               sync.Mutex{},
		replicaResponses: make(chan interface{}, ReplicaResponsesChannelSize),
		replicaID:        ReplicaID,
		lockMap:          make(map[string]int),
		slotIn:           1,
		slotOut:          1,
		requests:         make([]Command, 0),
		proposals:        make(map[int]Command),
		decisions:        make(map[int]Command),
		leaders:          LeaderAddresses,
		listener:         listener,
		dead:             0,
		Address:          listener.Addr().String(),
	}
	replica.newRequest = sync.Cond{L: &replica.mu}
	replica.somethingPerformed = sync.Cond{L: &replica.mu}
	server.Register(replica)

	go replica.propose()
	go replica.perform()

	go func() {
		for !replica.isDead() {
			connection, err := replica.listener.Accept()
			if err == nil {
				go server.ServeConn(connection)
			} else if err != nil && !replica.isDead() {
				log.Fatalf("Replica %d failed to accept connection, %s\n", ReplicaID, err)
			}
		}
	}()
	return replica
}
