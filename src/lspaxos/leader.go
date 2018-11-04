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
    LeaderInitialTimeout   = 1000
    LeaderAdditiveIncrease = 500
    LeaderMultDecrease     = 2
)

type Leader struct {
    // Identifies the leader as well as their round
    Ballot Ballot

    // Addresses of the acceptors
    Acceptors []string

    // Flag to identify if this leader is the commander
    Active bool

    // Number of responses to declare majority
    Majority int

    // Current timeout for Scout process
    Timeout int

    // Channel that communicates with acceptors in the Scout process
    ScoutChannel chan interface{}

    // Lock to control access to the proposals map
    mu sync.Mutex

    // Proposals map (slot number to request from a replica)
    Proposals map[int]Command
}

func (thisLeader *Leader) scout() {
    for {
        // Set of acceptors we've received from
        var received = make(map[int64]bool)
        // Probe the acceptors
        var request = ScoutRequest{Ballot: thisLeader.Ballot}
        for _, acceptor := range thisLeader.Acceptors {
            response := new(ScoutResponse)
            go Call(
                acceptor,
                "Acceptor.ExecutePropose",
                request,
                response,
                thisLeader.ScoutChannel,
            )
        }

        for len(received) < thisLeader.Majority {
            response := <- thisLeader.ScoutChannel
            if response == false {
                // Guaranteed to get a majority of responses according to spec
                // (i.e. Don't need to resend messages)
                log.Printf(
                    "Failed to get response from an acceptor on scout %d\n",
                    thisLeader.Ballot.Leader,
                )
                continue
            }

            var res = response.(ScoutResponse)
            
            // Merge the proposal map with what was received from
            // the acceptor
            thisLeader.mu.Lock()
            for slot := range res.AcceptedValues {
                thisLeader.Proposals[slot] = res.AcceptedValues[slot]
            }
            thisLeader.mu.Unlock()

            var compareResult = thisLeader.Ballot.Compare(res.Ballot)
            if compareResult< 0 {
                // We're pre-empted by somebody else, exit the loop
                break
            } else if compareResult == 0 {
                // Only record if the acceptor updated with our ballot number
                received[res.AcceptorID] = true
            }

            // Else case: It's a stale response from a previous round
            // If our ballot is higher than the one the acceptor responded
            // with, obviously this is old, because the acceptor will update
            // its ballot number immediately upon seeing a higher ballot
            // So just ignore
        }

        // Case where we got pre-empted
        if len(received) < thisLeader.Majority {
            // Sleep, then increment our timeout and ballot round
            // Make explicit that we are no longer the leader
            time.Sleep(time.Duration(thisLeader.Timeout) * time.Millisecond)
            thisLeader.Timeout += LeaderAdditiveIncrease
            thisLeader.Ballot.Number++
            thisLeader.Active = false
        } else {
            // Hooray, we are the leader
            // Decrement our timeout and then break out of the Scout loop
            thisLeader.Active = true
            thisLeader.Timeout /= LeaderMultDecrease
            break
        }
    }

    // TODO: How to start commander?
    // Is the commander already running?
    // (i.e. go thisLeader.commander())
    // If it's already running, we probably need a conditional variable
    // (i.e. the commander process waits until active == true and
    // there are proposals to execute).
    // However, PMMC implies that each slot spawns a commander,
    // so the thing to do here would be to async call a commander()
    // for each slot in the thisLeader.Proposals map.
}

func StartLeader(LeaderID int, Acceptors []string, Port string) (err error){
    thisLeader := &Leader{
        Ballot:       Ballot{Number: 0, Leader: LeaderID},
        Acceptors:    Acceptors,
        Active:       false,
        Majority:     (len(Acceptors) + 1) / 2 + (len(Acceptors) + 1) % 2,
        Timeout:      LeaderInitialTimeout,
        ScoutChannel: make(chan interface{}, len(Acceptors)),
        Proposals:    make(map[int]Command),
    }

    go thisLeader.scout()

    rpc.Register(thisLeader)
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