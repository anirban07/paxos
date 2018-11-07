package lspaxos

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strings"
	"time"
)

const (
	additiveIncrease       = 500
	multiplicativeDecrease = 2
)

type Client struct {
	// Unique identifier of the client
	clientID int

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
func StartClientWithSpec(
	ClientID int,
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

			var resp = response.(*ClientResponse)
			if resp.MsgID != thisClient.msgID {
				// Stale message
				continue
			}

			switch resp.Err {
			case ErrLockHeld:
				// Add constant amount to timeout
				// Wait, then resend commands
				// Have to increment the message id to deal with stale responses
				log.Printf("Client %d requested a held lock\n", thisClient.clientID)
				thisClient.timeoutMillis += additiveIncrease
				time.Sleep(time.Duration(thisClient.timeoutMillis) * time.Millisecond)
				log.Printf("Client %d woke up\n", thisClient.clientID)
				thisClient.msgID++
				command.MsgID = thisClient.msgID
				thisClient.SendCommand(command, done)
				continue
			case ErrInvalidUnlock:
				log.Printf("Client %d issued invalid unlock request %+v\n", thisClient.clientID, command)
			case OK:
				log.Printf("Client %d successfully executed %+v\n", thisClient.clientID, command)
			}

			// Either way, we're here if the lock didn't exist
			// or if the command succeeded. Need to exit and then decrease the
			// timeout by a factor  of multiplicativeDecrease
			thisClient.timeoutMillis /= multiplicativeDecrease
			break
		}

		// Start the next message
		thisClient.msgID++
	}
	return nil
}

func StartClient(ClientID int, Replicas []string) Client {
	return Client{
		clientID:      ClientID,
		msgID:         1,
		replicas:      Replicas,
		timeoutMillis: 100,
	}
}

func (thisClient *Client) Lock(LockName string) Err {
	command := Command{
		LockName: LockName,
		LockOp:   Lock,
		MsgID:    thisClient.msgID,
		ClientID: thisClient.clientID,
	}
	return thisClient.sendAndWaitForResponse(command)
}

func (thisClient *Client) Unlock(LockName string) Err {
	command := Command{
		LockName: LockName,
		LockOp:   Unlock,
		MsgID:    thisClient.msgID,
		ClientID: thisClient.clientID,
	}
	return thisClient.sendAndWaitForResponse(command)
}

func (thisClient *Client) ChanneledLock(LockName string, errChan chan Err) {
	command := Command{
		LockName: LockName,
		LockOp:   Lock,
		MsgID:    thisClient.msgID,
		ClientID: thisClient.clientID,
	}
	errChan <- thisClient.sendAndWaitForResponse(command)
}

func (thisClient *Client) ChanneledUnlock(LockName string, errChan chan Err) {
	command := Command{
		LockName: LockName,
		LockOp:   Unlock,
		MsgID:    thisClient.msgID,
		ClientID: thisClient.clientID,
	}
	errChan <- thisClient.sendAndWaitForResponse(command)
}

func (thisClient *Client) sendAndWaitForResponse(command Command) Err {
	// Send the command to each replica
	defer func() { thisClient.msgID++ }()
	done := make(chan interface{})
	var responseCount = 0
	thisClient.SendCommand(command, done)
	for {
		response := <-done
		responseCount++
		if response == false {
			if responseCount < len(thisClient.replicas) {
				log.Printf("Client %d connection error\n", thisClient.clientID)
				continue
			} else {
				return ErrConnectionError
			}
		}
		clientResponse := response.(*ClientResponse)
		if clientResponse.MsgID == thisClient.msgID {
			return clientResponse.Err
		}
	}
}

// Send a command to every replica asynchronously
func (thisClient *Client) SendCommand(Command Command, Done chan interface{}) {
	log.Printf("Client %d sent request %+v\n", thisClient.clientID, Command)
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
