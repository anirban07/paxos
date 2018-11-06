package lspaxos

import (
	"fmt"
	"testing"
	"time"
)

const (
	acceptorOffset int    = 10000
	leaderOffset   int    = 20000
	replicaOffset  int    = 30000
	ip             string = "localhost:"
)

func startAcceptors(
	numAcceptors int,
) (acceptorAddresses []string, acceptors []*Acceptor) {
	acceptorAddresses = make([]string, numAcceptors)
	acceptors = make([]*Acceptor, numAcceptors)
	for i := 0; i < numAcceptors; i++ {
		acceptor := StartAcceptor(i, "")
		acceptorAddresses[i] = acceptor.Address
		acceptors[i] = acceptor
	}
	return acceptorAddresses, acceptors
}

func startLeaders(
	numLeaders int,
	acceptors []string,
) (leaderAddresses []string, leaders []*Leader) {
	leaderAddresses = make([]string, numLeaders)
	leaders = make([]*Leader, numLeaders)
	for i := 0; i < numLeaders; i++ {
		leader := StartLeader(i, acceptors, "")
		leaderAddresses[i] = leader.Address
		leaders[i] = leader
	}
	return leaderAddresses, leaders
}

func startReplicas(
	numReplicas int,
	leaders []string,
) (replicaAddresses []string, replicas []*Replica) {
	replicaAddresses = make([]string, numReplicas)
	replicas = make([]*Replica, numReplicas)
	for i := 0; i < numReplicas; i++ {
		replica := StartReplica(i, leaders, "")
		replicaAddresses[i] = replica.Address
		replicas[i] = replica
	}
	return replicaAddresses, replicas
}

func failOnError(t *testing.T, err Err, format string, args ...interface{}) {
	if err != OK {
		t.Errorf(format, args)
	}
}

func cleanup(acceptors []*Acceptor, leaders []*Leader, replicas []*Replica) {
	for _, acceptor := range acceptors {
		acceptor.kill()
	}
	for _, leader := range leaders {
		leader.kill()
	}
	for _, replica := range replicas {
		replica.kill()
	}
}

func TestKillAcceptor(t *testing.T) {
	_, acceptors := startAcceptors(1)
	// time.Sleep(1 * time.Second)
	if acceptors[0].isDead() {
		t.Errorf("Acceptor is dead, expected to be alive\n")
	}
	acceptors[0].kill()
	if !acceptors[0].isDead() {
		t.Errorf("Acceptor is alive, expected to be dead\n")
	}

}

func TestKillLeader(t *testing.T) {
	acceptorAddresses, acceptors := startAcceptors(1)
	_, leaders := startLeaders(1, acceptorAddresses)
	if leaders[0].isDead() {
		t.Errorf("Leader is dead, expected to be alive\n")
	}
	if acceptors[0].isDead() {
		t.Errorf("Acceptor is dead, expected to be alive\n")
	}
	acceptors[0].kill()
	leaders[0].kill()
	if !leaders[0].isDead() {
		t.Errorf("Leader is alive, expected to be dead\n")
	}
	if !acceptors[0].isDead() {
		t.Errorf("Acceptor is alive, expected to be dead\n")
	}
}

func TestKillReplicas(t *testing.T) {
	acceptorAddresses, acceptors := startAcceptors(1)
	leaderAddresses, leaders := startLeaders(1, acceptorAddresses)
	_, replicas := startReplicas(1, leaderAddresses)
	if replicas[0].isDead() {
		t.Errorf("Replica is dead, expected to be alive\n")
	}
	if leaders[0].isDead() {
		t.Errorf("Leader is dead, expected to be alive\n")
	}
	if acceptors[0].isDead() {
		t.Errorf("Acceptor is dead, expected to be alive\n")
	}
	acceptors[0].kill()
	leaders[0].kill()
	replicas[0].kill()
	if !replicas[0].isDead() {
		t.Errorf("Replica is alive, expected to be dead\n")
	}
	if !leaders[0].isDead() {
		t.Errorf("Leader is alive, expected to be dead\n")
	}
	if !acceptors[0].isDead() {
		t.Errorf("Acceptor is alive, expected to be dead\n")
	}
}

func TestBasic(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)

	time.Sleep(2 * time.Second)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	t.Logf("Grabbed lockA\n")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	time.Sleep(5 * time.Second)
	cleanup(acceptors, leaders, replicas)
	fmt.Printf("  ... Passed\n")
}
