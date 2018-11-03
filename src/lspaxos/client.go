package lspaxos

import (
	"bufio"
	"log"
	"os"
)

type Client struct {
	// Unique identifier of the client
	ClientID int

	// Initial request sequence number / message id number
	// Should start at 0
	MsgID int
}

// Starts a client that issues requests in the order
// of the given specification of lock/unlock operations
func StartClient(Spec []string) (err error) {
	return nil
}

func ReadSpec(Filename string) []string {
	fd, err := os.Open(Filename)
	if err != nil {
		log.Println("Error opening file " + Filename)
		return nil
	}

	reader := bufio.NewReader(fd)
	spec := make([]string, 0)
	for err != nil {
		line, _ := reader.ReadString('\n')
		spec = append(spec, line)
	}

	return spec
}
