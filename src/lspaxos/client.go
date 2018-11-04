package lspaxos

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
	"time"
)

type Client struct {
	// Unique identifier of the client
	clientID int64

	// Initial request sequence number / message id number
	// Should start at 0
	msgID int

	// Addresses of the replica servers
	replicas []string

	// Current time out
	timeoutMillis int
}

// Starts a client that issues requests in the order
// of the given specification of lock/unlock operations
func StartClient(
	ClientID int64,
	Replicas []string,
	Spec []string,
) (err error) {
	thisClient := Client{
		clientID:      ClientID,
		msgID:         1,
		replicas:      Replicas,
		timeoutMillis: 0,
	}
	done := make(chan interface{}, len(Spec))
	for _, line := range Spec {
		parts := strings.Fields(line)
		command := Command{LockName: parts[1],
			LockOp:   LockOp(parts[0]),
			MsgID:    thisClient.msgID,
			ClientID: thisClient.clientID}

		// Send the command to each replica
		thisClient.SendCommand(command, done)

		// Grab responses, break on the first valid one.
		for {
			response := <-done
			if response == false {
				log.Println("Error occurred in the RPC, break")
				return errors.New("Failed to call replica, invariant broken")
			}

			var resp = response.(ClientResponse)
			if resp.MsgID != thisClient.msgID {
				// Stale message
				continue
			}

			if resp.Err == ErrLockHeld {
				// Add constant amount to timeout
				// Wait, then resend commands
				// Have to increment the message id to deal with stale responses
				thisClient.timeoutMillis += 500
				time.Sleep(time.Duration(thisClient.timeoutMillis) * time.Millisecond)
				thisClient.msgID++
				thisClient.SendCommand(command, done)
				continue
			} else if resp.Err == OK {
				log.Printf("Client %d successfully executed %+v\n", thisClient.clientID, command)
			}

			// Either way, we're here if the lock didn't exist
			// or if the command succeeded. Need to exit and then decrease the
			// timeout by a factor  of 2
			thisClient.timeoutMillis /= 2
			break
		}

		// Start the next message
		thisClient.msgID++
	}
	close(done)
	return nil
}

// Send a command to every replica asynchronously
func (thisClient *Client) SendCommand(Command Command, Done chan interface{}) {
	for _, server := range thisClient.replicas {
		request := ClientRequest{Command: Command}
		response := new(ClientResponse)
		go Call(server, "Replica.ExecuteRequest", request, response, Done)
	}
}

// Read the specification for the client from a file
func ReadSpec(Filename string) []string {
	fd, err := os.Open(Filename)
	if err != nil {
		log.Println("Error opening file " + Filename)
		panic(err)
	}
	defer fd.Close()

	reader := bufio.NewScanner(fd)
	spec := make([]string, 0)

	for reader.Scan() {
		spec = append(spec, reader.Text())
	}

	return spec
}
