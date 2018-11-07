package lspaxos

import (
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
	acceptorAddresses []string,
) (leaderAddresses []string, leaders []*Leader) {
	leaderAddresses = make([]string, numLeaders)
	leaders = make([]*Leader, numLeaders)
	for i := 0; i < numLeaders; i++ {
		leader := StartLeader(i, acceptorAddresses, "")
		leaderAddresses[i] = leader.Address
		leaders[i] = leader
	}
	return leaderAddresses, leaders
}

func startReplicas(
	numReplicas int,
	leaderAddresses []string,
) (replicaAddresses []string, replicas []*Replica) {
	replicaAddresses = make([]string, numReplicas)
	replicas = make([]*Replica, numReplicas)
	for i := 0; i < numReplicas; i++ {
		replica := StartReplica(i, leaderAddresses, "")
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
		if !acceptor.isDead() {
			acceptor.kill()
		}
	}
	for _, leader := range leaders {
		if !leader.isDead() {
			leader.kill()
		}
	}

	for _, replica := range replicas {
		if !replica.isDead() {
			replica.kill()
		}
	}
}

func TestKillAcceptor(t *testing.T) {
	_, acceptors := startAcceptors(1)
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

func Test1c1r1l3a(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c1r3l3a(t *testing.T) {
	numReplicas := 1
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r1l3a(t *testing.T) {
	numReplicas := 3
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r3l3a(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r3l3aFailingReplicas(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	replicas[0].kill()
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	replicas[2].kill()
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r3l3aFailingLeaders(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	leaders[1].kill()
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	leaders[2].kill()
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r3l3aFailingAcceptors(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	acceptors[1].kill()
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test1c3r3l3aFailingAll(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	err := client0.Lock(lockA)
	failOnError(t, err, "")
	leaders[1].kill()
	replicas[0].kill()
	err = client0.Lock(lockB)
	failOnError(t, err, "")
	acceptors[0].kill()
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	leaders[2].kill()
	replicas[2].kill()
	err = client0.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test2c1r1l3aIndependentLocks(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses)
	client1 := StartClient(1, replicaAddresses)

	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client1.Lock(lockB)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client1.Unlock(lockB)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test2c1r1l3aContendingLocks(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := startAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := startLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := startReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	client0 := StartClient(0, replicaAddresses)
	client1 := StartClient(1, replicaAddresses)

	err := client0.Lock(lockA)
	failOnError(t, err, "")
	err = client1.Lock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	err = client1.Unlock(lockA)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)

}
