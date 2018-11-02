package lspaxos

import (
    "bufio"
    "fmt"
    "os"
)

type Client struct {
    // Unique identifier of the client
    ClientId int
     
    // Initial request sequence number / message id number
    // Should start at 0
    MsgId int
}


// Starts a client that issues requests in the order
// of the given specification of lock/unlock operations
func StartClient(Spec []string) (err error) {
    return
}

func ReadSpec(Filename string) []string {
    fd, err := os.Open(Filename)
    if err != nil {
        fmt.Println("Error opening file " + Filename)
        return nil
    }

    reader := bufio.NewReader(fd)
    spec := make([]string, 0)
    for err != nil {
        line, err := reader.ReadString('\n')
        spec = append(spec, line)
    }

    return spec
}