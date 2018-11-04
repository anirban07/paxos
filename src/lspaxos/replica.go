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
	replicaResponses chan interface{}

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

	// Condition variable for new item added to Requests
	newRequest sync.Cond

	// Client requests not yet proposed or decided
	requests []ClientRequest

	// Proposals that are currently outstanding
	proposals map[int]ClientRequest

	// Proposals that are known to have been decided
	decisions map[int]ClientRequest

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
		for _, clientRequest := range thisReplica.requests {
			request := ReplicaRequest{
				Command: clientRequest.Command,
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
			thisReplica.proposals[thisReplica.slotIn] = clientRequest
			thisReplica.slotIn++
		}
		thisReplica.requests = nil
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
		requests:         make([]ClientRequest, 8),
		proposals:        make(map[int]ClientRequest),
		decisions:        make(map[int]ClientRequest),
		leaders:          Leaders,
	}
	thisReplica.newRequest = sync.Cond{L: &thisReplica.mu}
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
	thisReplica.requests = append(thisReplica.requests, req)
	thisReplica.mu.Unlock()
	thisReplica.newRequest.Signal()

	for {

	}
}
