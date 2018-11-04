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
	// Mutex to synchronize access to this struct
	mu sync.Mutex

	// Channel ReplicaResponse from Leader
	ReplicaResponses chan interface{}

	// Unique identifier of the replica
	ReplicaID int64

	// Map from lock name to client holding it
	LockMap map[string]int64

	// The index of the next slot in which the replica
	// has not yet proposed any command, initially 1
	SlotIn int

	// The index of the next slot for which it needs to learn a
	// decision before it can update its copy of the application state
	SlotOut int

	// Condition variable for new item added to Requests
	NewRequest sync.Cond

	// Client requests not yet proposed or decided
	Requests []ClientRequest

	// Proposals that are currently outstanding
	Proposals map[int]ClientRequest

	// Proposals that are known to have been decided
	Decisions map[int]ClientRequest

	// Leaders
	Leaders []string
}

func (thisReplica *Replica) propose() {
	for {
		thisReplica.mu.Lock()
		for len(thisReplica.Requests) == 0 {
			thisReplica.NewRequest.Wait()
		}
		// Propose values. Should empty out the Requests
		for _, clientRequest := range thisReplica.Requests {
			request := ReplicaRequest{
				Command: clientRequest.Command,
				Slot:    thisReplica.SlotIn,
			}
			for _, leader := range thisReplica.Leaders {
				response := new(ReplicaResponse)
				go Call(
					leader,
					"Leader.ExecutePropose",
					request,
					response,
					thisReplica.ReplicaResponses,
				)
			}
			thisReplica.Proposals[thisReplica.SlotIn] = clientRequest
			thisReplica.SlotIn++
		}
		thisReplica.Requests = nil
		thisReplica.mu.Unlock()
	}
}

func StartReplica(ReplicaID int64, Leaders []string, Port string) (err error) {
	thisReplica := &Replica{
		mu:               sync.Mutex{},
		ReplicaResponses: make(chan interface{}, ReplicaResponsesChannelSize),
		ReplicaID:        ReplicaID,
		LockMap:          make(map[string]int64),
		SlotIn:           1,
		SlotOut:          1,
		Requests:         make([]ClientRequest, 8),
		Proposals:        make(map[int]ClientRequest),
		Decisions:        make(map[int]ClientRequest),
		Leaders:          Leaders,
	}
	thisReplica.NewRequest = sync.Cond{L: &thisReplica.mu}
	rpc.Register(thisReplica)
	log.Printf("Replica %d listening on port %s\n", ReplicaID, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()
	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on " + string(ReplicaID))
		return err
	}

	go thisReplica.propose()
	rpc.Accept(listener)
	return nil
}

func (thisReplica *Replica) ExecuteRequest(req ClientRequest, res *ClientResponse) (err error) {
	// TODO: Check for duplicate req in Requests, Proposals, Decisions
	thisReplica.mu.Lock()
	thisReplica.Requests = append(thisReplica.Requests, req)
	thisReplica.mu.Unlock()
	thisReplica.NewRequest.Signal()

	for {

	}
	return nil
}
