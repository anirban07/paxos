package lspaxos

import (
	"testing"
	"time"
)

const (
	acceptorOffset       int    = 10000
	leaderOffset         int    = 20000
	replicaOffset        int    = 30000
	ip                   string = "localhost:"
	timeoutMillis        int    = 100
	timeoutMillisAddInc  int    = 50
	timeoutMillisMultDec int    = 2
)

func failOnError(t *testing.T, err Err, format string, args ...interface{}) {
	if err != OK {
		t.Errorf(string(err))
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
	_, acceptors := StartAcceptors(1)
	if acceptors[0].isDead() {
		t.Errorf("Acceptor is dead, expected to be alive\n")
	}
	acceptors[0].kill()
	if !acceptors[0].isDead() {
		t.Errorf("Acceptor is alive, expected to be dead\n")
	}

}

func TestKillLeader(t *testing.T) {
	acceptorAddresses, acceptors := StartAcceptors(1)
	_, leaders := StartLeaders(1, acceptorAddresses)
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
	acceptorAddresses, acceptors := StartAcceptors(1)
	leaderAddresses, leaders := StartLeaders(1, acceptorAddresses)
	_, replicas := StartReplicas(1, leaderAddresses)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	replicas[0].kill()
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	leaders[1].kill()
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	leaders[1].kill()
	replicas[0].kill()
	err = client0.TryLock(lockB)
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

func Test1c1r1l3aInvalidLockRequest(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)

	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client0.Unlock(lockB)
	if err != ErrInvalidUnlock {
		t.Fatalf("Should have been invalid unlock, %s\n", err)
	}

	err = client0.Unlock(lockA)
	failOnError(t, err, "")
	cleanup(acceptors, leaders, replicas)
}

func Test2c1r1l3aIndependentLocks(t *testing.T) {
	numReplicas := 1
	numLeaders := 1
	numAcceptors := 3

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	lockB := "B"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	client1 := StartClient(1, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)

	err := client0.TryLock(lockA)
	failOnError(t, err, "")
	err = client1.TryLock(lockB)
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

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	client1 := StartClient(1, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)

	client0Channel := make(chan Err, 100)
	client1Channel := make(chan Err, 100)

	client0.ChanneledLock(lockA, client0Channel)
	// Let client0 grab the lock
	err := <-client0Channel
	failOnError(t, err, "")

	done := make(chan bool, 5)
	go func() {
		var err Err = ErrLockHeld
		sleepTimeMillis := 100
		sleepTimeIncrement := 10
		for ; err == ErrLockHeld; err = <-client1Channel {
			time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
			sleepTimeMillis += sleepTimeIncrement
			go client1.ChanneledLock(lockA, client1Channel)
		}
		done <- true
	}()
	// This will cause a few ErrLockHeld
	time.Sleep(1 * time.Second)

	go client0.ChanneledUnlock(lockA, client0Channel)

	// Wait for client1 to grab the lock
	<-done
	client1.ChanneledUnlock(lockA, client1Channel)
	cleanup(acceptors, leaders, replicas)

}

func Test2c1r1l3aContendingLocksFailingAll(t *testing.T) {
	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, acceptors := StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, leaders := StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, replicas := StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	lockA := "A"
	client0 := StartClient(0, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)
	client1 := StartClient(1, replicaAddresses, timeoutMillis, timeoutMillisAddInc, timeoutMillisMultDec)

	client0Channel := make(chan Err, 100)
	client1Channel := make(chan Err, 100)

	client0.ChanneledLock(lockA, client0Channel)
	// Let client0 grab the lock
	err := <-client0Channel
	failOnError(t, err, "")
	replicas[0].kill()
	leaders[2].kill()

	done := make(chan bool, 5)
	go func() {
		var err Err = ErrLockHeld
		sleepTimeMillis := 100
		sleepTimeIncrement := 10
		for ; err == ErrLockHeld; err = <-client1Channel {
			time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
			sleepTimeMillis += sleepTimeIncrement
			go client1.ChanneledLock(lockA, client1Channel)
		}
		done <- true
	}()
	// This will cause a few ErrLockHeld
	time.Sleep(1 * time.Second)

	leaders[1].kill()
	go client0.ChanneledUnlock(lockA, client0Channel)
	acceptors[1].kill()
	// Wait for client1 to grab the lock
	<-done
	replicas[2].kill()
	client1.ChanneledUnlock(lockA, client1Channel)
	cleanup(acceptors, leaders, replicas)
}
