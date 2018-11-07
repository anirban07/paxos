package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"lspaxos"
	"strings"
	"sync"
	"time"
)

const (
	ip = "localhost:"
)

func main() {
	log.SetOutput(ioutil.Discard)
	lock := "lock"

	numReplicas := 3
	numLeaders := 3
	numAcceptors := 3

	acceptorAddresses, _ := lspaxos.StartAcceptors(numAcceptors)
	time.Sleep(500 * time.Millisecond)
	leaderAddresses, _ := lspaxos.StartLeaders(numLeaders, acceptorAddresses)
	replicaAddresses, _ := lspaxos.StartReplicas(numReplicas, leaderAddresses)
	time.Sleep(500 * time.Millisecond)

	client0 := lspaxos.StartClient(0, replicaAddresses, 100, 50, 2)
	client1 := lspaxos.StartClient(1, replicaAddresses, 100, 50, 2)

	fmt.Println("client0 trying to grab lock")
	client0.Lock(lock)
	fmt.Println("client0 grabbed the lock")
	go func() {
		fmt.Println("client1 trying to grab lock")
		client1.Lock(lock)
		fmt.Println("client1 grabbed the lock")
	}()
	time.Sleep(5 * time.Second)
	client0.Unlock(lock)
	client1.Unlock(lock)
}

func test() {
	AcceptorAddrs := []string{ip + "10000", ip + "10001", ip + "10002"}
	LeaderAddrs := []string{ip + "20000", ip + "20001"}
	ReplicaAddrs := []string{ip + "30000"}

	// Start the acceptors
	for id, addr := range AcceptorAddrs {
		go lspaxos.StartAcceptor(id, ":"+strings.Split(addr, ":")[1])
	}

	// Start the leaders
	for id, addr := range LeaderAddrs {
		go lspaxos.StartLeader(id, AcceptorAddrs, ":"+strings.Split(addr, ":")[1])
	}

	// Start the replicas
	for id, addr := range ReplicaAddrs {
		go lspaxos.StartReplica(id, LeaderAddrs, ":"+strings.Split(addr, ":")[1])
	}
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	Spec := lspaxos.ReadSpec("src/specs/test_spec.txt")
	wg.Add(2)
	go func() {
		lspaxos.StartClientWithSpec(666, ReplicaAddrs, Spec)
		wg.Done()
	}()

	go func() {
		lspaxos.StartClientWithSpec(777, ReplicaAddrs, Spec)
		wg.Done()
	}()

	wg.Wait()
}
