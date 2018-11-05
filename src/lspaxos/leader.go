package lspaxos

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

const (
	leaderInitialTimeout   = 1000
	leaderAdditiveIncrease = 500
	leaderMultDecrease     = 2
)

type Leader struct {
	// Identifies the leader as well as their round
	ballot Ballot

	// Addresses of the acceptors
	acceptors []string

	// Flag to identify if this leader is the commander
	active bool

	// Number of responses to declare majority
	majority int

	// Current timeout for Scout process
	timeout int

	// Channel that communicates with acceptors in the Scout process
	scoutChannel chan interface{}

	// Lock to control access to the proposals map
	mu sync.Mutex

	// Proposals map (slot number to command)
	proposals map[int]Command

	// Decisions map (slot number to command)
	decisions map[int]Command

	// Condition variable for when you need to scout
	needToScout sync.Cond

	// Condition variable for when something is decided
	somethingDecided sync.Cond
}

func (thisLeader *Leader) scout() {
	for {
		thisLeader.mu.Lock()
		for thisLeader.active {
			thisLeader.needToScout.Wait()
		}

		for !thisLeader.active {
			// Set of acceptors we've received from
			var received = make(map[int64]bool)
			// Probe the acceptors
			var request = ScoutRequest{Ballot: thisLeader.ballot}
			for _, acceptor := range thisLeader.acceptors {
				response := new(ScoutResponse)
				go Call(
					acceptor,
					"Acceptor.ExecutePropose",
					request,
					response,
					thisLeader.scoutChannel,
				)
			}

			for len(received) < thisLeader.majority {
				response := <-thisLeader.scoutChannel
				if response == false {
					// Guaranteed to get a majority of responses according to spec
					// (i.e. Don't need to resend messages)
					log.Printf(
						"Failed to get response from an acceptor on scout %d\n",
						thisLeader.ballot.Leader,
					)
					continue
				}

				var res = response.(ScoutResponse)
				var compareResult = thisLeader.ballot.Compare(res.Ballot)
				if compareResult < 0 {
					// We're pre-empted by somebody else, exit the loop
					break
				} else if compareResult == 0 {
					// Only record if the acceptor updated with our ballot number
					received[res.AcceptorID] = true
					// Merge the proposal map with what was received from
					// the acceptor
					for slot, acceptedCommand := range res.AcceptedValues {
						thisLeader.proposals[slot] = acceptedCommand
					}
				}

				// Else case: It's a stale response from a previous round
				// If our ballot is higher than the one the acceptor responded
				// with, obviously this is old, because the acceptor will update
				// its ballot number immediately upon seeing a higher ballot
				// So just ignore
			}

			// Case where we got pre-empted
			if len(received) < thisLeader.majority {
				// Sleep, then increment our timeout and ballot round
				// Make explicit that we are no longer the leader
				time.Sleep(time.Duration(thisLeader.timeout) * time.Millisecond)
				thisLeader.timeout += leaderAdditiveIncrease
				thisLeader.ballot.Number++
				thisLeader.active = false
			} else {
				// Hooray, we are the leader
				// Decrement our timeout and then break out of the Scout loop
				thisLeader.active = true
				thisLeader.timeout /= leaderMultDecrease
			}
		}

		for slot, command := range thisLeader.proposals {
			go thisLeader.commander(slot, command, thisLeader.ballot)
		}
		thisLeader.mu.Unlock()
	}
}

func (thisLeader *Leader) commander(slot int, command Command, ballot Ballot) {
	// Set of acceptors we've received from
	var received = make(map[int64]bool)

	// Channel that communicates with acceptors in the Scout process
	commanderChannel := make(chan interface{}, len(thisLeader.acceptors))
	defer close(commanderChannel)

	// Probe the acceptors
	var request = CommanderRequest{Command: command, Slot: slot, Ballot: ballot}
	for _, acceptor := range thisLeader.acceptors {
		response := new(CommanderResponse)
		go Call(
			acceptor,
			"Acceptor.ExecuteAccept",
			request,
			response,
			commanderChannel,
		)
	}

	for len(received) < thisLeader.majority {
		response := <-commanderChannel
		if response == false {
			log.Printf("Response from acceptor failed on scout %d\n", thisLeader.ballot.Leader)
			continue
		}
		commanderResponse := response.(CommanderResponse)
		if ballot.Compare(commanderResponse.Ballot) == 0 {
			received[commanderResponse.AcceptorID] = true
		} else if ballot.Compare(commanderResponse.Ballot) < 0 {
			// Preempted
			thisLeader.mu.Lock()
			if thisLeader.ballot.Compare(ballot) == 0 {
				thisLeader.active = false
				thisLeader.needToScout.Signal()
			}
			thisLeader.mu.Unlock()
			return
		}
	}

	thisLeader.mu.Lock()
	thisLeader.decisions[slot] = command
	delete(thisLeader.proposals, slot)
	thisLeader.somethingDecided.Broadcast()
	thisLeader.mu.Unlock()
}

func (thisLeader *Leader) ExecutePropose(req ReplicaRequest, res *ReplicaResponse) (err error) {
	res.Slot = req.Slot
	thisLeader.mu.Lock()
	defer thisLeader.mu.Unlock()
	decision, decided := thisLeader.decisions[req.Slot]
	if decided {
		res.Command = decision
	} else {
		alreadyProposed := false
		for _, proposal := range thisLeader.proposals {
			if req.Command.Equals(proposal) {
				alreadyProposed = true
				break
			}
		}
		_, reqSlotIsUsed := thisLeader.proposals[req.Slot]
		if !alreadyProposed && !reqSlotIsUsed {
			thisLeader.proposals[req.Slot] = req.Command
			if thisLeader.active {
				go thisLeader.commander(req.Slot, req.Command, thisLeader.ballot)
			}
		}

		for !decided {
			thisLeader.somethingDecided.Wait()
		}
		res.Command = thisLeader.decisions[req.Slot]
	}
	return nil
}

func StartLeader(LeaderID int, Acceptors []string, Port string) (err error) {
	thisLeader := &Leader{
		ballot:       Ballot{Number: 0, Leader: LeaderID},
		acceptors:    Acceptors,
		active:       false,
		majority:     (len(Acceptors)+1)/2 + (len(Acceptors)+1)%2,
		timeout:      leaderInitialTimeout,
		scoutChannel: make(chan interface{}, ChannelBufferSize),
		proposals:    make(map[int]Command),
		decisions:    make(map[int]Command),
	}
	thisLeader.needToScout = sync.Cond{L: &thisLeader.mu}
	thisLeader.somethingDecided = sync.Cond{L: &thisLeader.mu}
	rpc.Register(thisLeader)

	go thisLeader.scout()

	log.Printf("Leader %d listening on port %s\n", LeaderID, Port)
	listener, err := net.Listen("tcp", ":"+Port)
	defer listener.Close()
	if err != nil {
		err = errors.New("Failed to set up listening port " + Port + " on leader " + string(LeaderID))
		return err
	}

	rpc.Accept(listener)
	return nil
}
