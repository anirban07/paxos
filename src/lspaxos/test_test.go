package lspaxos

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

const (
	acceptorOffset int    = 10000
	leaderOffset   int    = 20000
	replicaOffset  int    = 30000
	ip             string = "localhost:"
)

func startAcceptors(numAcceptors int) (acceptors []string) {
	acceptors = make([]string, numAcceptors)
	for i := 0; i < numAcceptors; i++ {
		port := strconv.Itoa(acceptorOffset + i)
		acceptors[i] = ip + port
		go StartAcceptor(i, port)
	}
	return acceptors
}

func startLeaders(numLeaders int, acceptors []string) (leaders []string) {
	leaders = make([]string, numLeaders)
	for i := 0; i < numLeaders; i++ {
		port := strconv.Itoa(leaderOffset + i)
		leaders[i] = ip + port
		go StartLeader(i, acceptors, port)
	}
	return leaders
}

func startReplicas(numReplicas int, leaders []string) (replicas []string) {
	replicas = make([]string, numReplicas)
	for i := 0; i < numReplicas; i++ {
		port := strconv.Itoa(replicaOffset + i)
		replicas[i] = ip + port
		go StartReplica(i, leaders, port)
	}
	return replicas
}

func failOnError(t *testing.T, err Err, format string, args ...interface{}) {
	if err != OK {
		t.Errorf(format, args)
	}
}

func TestBasic(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptors := startAcceptors(numAcceptors)
	leaders := startLeaders(numLeaders, acceptors)
	replicas := startReplicas(numReplicas, leaders)

	time.Sleep(2 * time.Second)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicas)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	t.Logf("Grabbed lockA\n")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	fmt.Printf("  ... Passed\n")
}
