package lspaxos

import (
	"log"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"
)

const (
	leaderInitialTimeout   = 100
	leaderAdditiveIncrease = 50
	leaderMultDecrease     = 2
)

type Leader struct {
	// Identifies the leader
	leaderID int

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

	// Listener
	listener net.Listener

	// Address
	Address string

	// For debugging
	dead int32
}

func (thisLeader *Leader) scout() {
	for {
		log.Printf("Leader %d scout trying to grab lock\n", thisLeader.leaderID)
		thisLeader.mu.Lock()
		for thisLeader.active {
			log.Printf("Leader %d is the leader, scout waiting...", thisLeader.leaderID)
			thisLeader.needToScout.Wait()
			log.Printf("Leader %d scout got woken up\n", thisLeader.leaderID)
		}

		for !thisLeader.active {
			// Set of acceptors we've received from
			var received = make(map[int]bool)
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

			// Listen for responses
			for len(received) < thisLeader.majority {
				response := <-thisLeader.scoutChannel
				if response == false {
					// Guaranteed to get a majority of responses according to spec
					// (i.e. Don't need to resend messages)
					log.Printf(
						"Failed to get response from an acceptor on scout %d\n",
						thisLeader.leaderID,
					)
					continue
				}

				var res = response.(*ScoutResponse)
				log.Printf("Leader %d got a scout response from acceptor %+v\n", thisLeader.leaderID, res)
				var compareResult = thisLeader.ballot.Compare(res.Ballot)
				if compareResult < 0 {
					// We're pre-empted by somebody else, exit the loop
					thisLeader.ballot.Number = res.Ballot.Number + 1
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
				log.Printf("Leader %d is sleeping for %d milliseconds\n", thisLeader.leaderID, thisLeader.timeout)
				time.Sleep(time.Duration(thisLeader.timeout) * time.Millisecond)
				log.Printf("Leader %d is done sleeping\n", thisLeader.leaderID)
				thisLeader.timeout += leaderAdditiveIncrease
				thisLeader.active = false
			} else {
				// Hooray, we are the leader
				// Decrement our timeout and then break out of the Scout loop
				thisLeader.active = true
				thisLeader.timeout /= leaderMultDecrease
				log.Printf("Leader %+v is Spartacus\n", thisLeader.ballot)
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
	var received = make(map[int]bool)

	// Channel that communicates with acceptors in the Scout process
	commanderChannel := make(chan interface{}, len(thisLeader.acceptors))

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
			log.Printf("Response from acceptor failed on scout %d\n", thisLeader.leaderID)
			continue
		}
		commanderResponse := response.(*CommanderResponse)
		log.Printf("Leader %d received a commander response from acceptor %+v, request %+v\n", thisLeader.leaderID, commanderResponse, request.Command)
		if ballot.Compare(commanderResponse.Ballot) == 0 {
			received[commanderResponse.AcceptorID] = true
		} else if ballot.Compare(commanderResponse.Ballot) < 0 {
			// Preempted
			thisLeader.mu.Lock()
			if thisLeader.ballot.Compare(ballot) == 0 {
				log.Printf("Leader %d is sleeping for %d milliseconds\n", thisLeader.leaderID, thisLeader.timeout)
				time.Sleep(time.Duration(thisLeader.timeout) * time.Millisecond)
				log.Printf("Leader %d is done sleeping\n", thisLeader.leaderID)
				thisLeader.timeout += leaderAdditiveIncrease
				thisLeader.active = false
				thisLeader.ballot.Number = commanderResponse.Ballot.Number + 1
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
	log.Printf("Leader %d got a replica request %+v\n", thisLeader.leaderID, req)
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

		for ; !decided; _, decided = thisLeader.decisions[req.Slot] {
			thisLeader.somethingDecided.Wait()

		}
		res.Command = thisLeader.decisions[req.Slot]
		log.Printf("Leader %d decided %+v for slot %d\n", thisLeader.leaderID, res.Command, res.Slot)
	}
	return nil
}

func (thisLeader *Leader) kill() {
	log.Printf("Killing leader %d\n", thisLeader.leaderID)
	atomic.StoreInt32(&thisLeader.dead, 1)
	if thisLeader.listener != nil {
		thisLeader.listener.Close()
	}
}

func (thisLeader *Leader) isDead() bool {
	return atomic.LoadInt32(&thisLeader.dead) != 0
}

//StartLeader starts an acceptor instance and returns an Leader struct.
//The struct can be used to kill this instance.
func StartLeader(LeaderID int, Acceptors []string, Address string) (leader *Leader) {
	server := rpc.NewServer()
	listener, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatalf(
			"Leader %d failed to set up listening address %s, %s\n",
			LeaderID,
			Address,
			err,
		)
		return nil
	}
	leader = &Leader{
		mu:           sync.Mutex{},
		leaderID:     LeaderID,
		ballot:       Ballot{Number: 0, Leader: LeaderID},
		acceptors:    Acceptors,
		active:       false,
		majority:     (len(Acceptors)+1)/2 + (len(Acceptors)+1)%2,
		timeout:      leaderInitialTimeout,
		scoutChannel: make(chan interface{}, ChannelBufferSize),
		proposals:    make(map[int]Command),
		decisions:    make(map[int]Command),
		listener:     listener,
		dead:         0,
		Address:      listener.Addr().String(),
	}
	leader.needToScout = sync.Cond{L: &leader.mu}
	leader.somethingDecided = sync.Cond{L: &leader.mu}
	server.Register(leader)

	go func() {
		for !leader.isDead() {
			log.Printf("Leader %d listening for requests\n", LeaderID)
			connection, err := leader.listener.Accept()
			if err == nil {
				log.Printf("Leader accepted request\n")
				go server.ServeConn(connection)
			} else {
				log.Fatalf("Leader %d failed to accept connection, %s\n", LeaderID, err)
			}
		}
	}()
	return leader
}
