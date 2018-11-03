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
	ClientId int64

	// Initial request sequence number / message id number
	// Should start at 0
	MsgId int

	// Addresses of the replica servers
	Replicas []string

	// Current time out
	TimeoutMillis int
}

// Starts a client that issues requests in the order
// of the given specification of lock/unlock operations
func (c *Client) StartClient(Spec []string) {
	c.MsgId = 1
	c.TimeoutMillis = 0
	done := make(chan interface{}, len(Spec))
	for _, line := range Spec {
		parts := strings.Fields(line)
		command := Command{LockName: parts[1],
			LockOp:   LockOp(parts[0]),
			MsgId:    c.MsgId,
			ClientId: c.ClientId}

		// Send the command to each replica
		c.SendCommand(command, done)

		// Grab responses, break on the first valid one.
		for {
			response := <-done
			if response == false {
				log.Println("Error occurred in the RPC, break")
				panic(errors.New("Failed to call replica, invariant broken"))
			}

			var resp = response.(ClientResponse)
			if resp.MsgId != c.MsgId {
				// Stale message
				continue
			}

			if resp.Err == ErrLockHeld {
				// Add constant amount to timeout
				// Wait, then resend commands
				// Have to increment the message id to deal with stale responses
				c.TimeoutMillis += 500
				time.Sleep(time.Duration(c.TimeoutMillis) * time.Millisecond)
				c.MsgId++
				c.SendCommand(command, done)
				continue
			} else if resp.Err == OK {
				log.Printf("Client %d successfully executed %+v\n", c.ClientId, command)
			}

			// Either way, we're here if the lock didn't exist
			// or if the command succeeded. Need to exit and then decrease the
			// timeout by a factor  of 2
			c.TimeoutMillis /= 2
			break
		}

		// Start the next message
		c.MsgId++
	}
	close(done)
}

// Send a command to every replica asynchronously
func (c *Client) SendCommand(Command Command, Done chan interface{}) {
	for _, server := range c.Replicas {
		request := ClientRequest{Command: Command}
		response := new(ClientResponse)
		go Call(server, "ReplicaServer.ExecuteRequest", request, response, Done)
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
