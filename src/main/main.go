package main

import (
	"lspaxos"
	"strings"
	"time"
)

const (
	ip = "localhost:"
)

func main() {
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

	Spec := lspaxos.ReadSpec("src/specs/test_spec.txt")
	go lspaxos.StartClientWithSpec(666, ReplicaAddrs, Spec)
	go lspaxos.StartClientWithSpec(777, ReplicaAddrs, Spec)

	for {
		time.Sleep(1 * time.Second)
	}
}
