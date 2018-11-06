package main

import (
	"lspaxos"
	"strings"
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
		go lspaxos.StartAcceptor(int64(id), strings.Split(addr, ":")[1])
	}

	// Start the leaders
	for id, addr := range LeaderAddrs {
		go lspaxos.StartLeader(id, AcceptorAddrs, strings.Split(addr, ":")[1])
	}

	// Start the replicas
	for id, addr := range ReplicaAddrs {
		go lspaxos.StartReplica(int64(id), LeaderAddrs, strings.Split(addr, ":")[1])
	}

	Spec := lspaxos.ReadSpec("src/specs/test_spec.txt")
	lspaxos.StartClient(int64(666), ReplicaAddrs, Spec)
}
