package main

import (
	"lspaxos"
)

func main() {
<<<<<<< HEAD
    lspaxos.StartAcceptorServer("Alpha", os.Args[1])
	//TestRpcServer(os.Args[1])
	/*var commands = lspaxos.ReadSpec("../specs/test_spec.txt")
	replicas := [1]string {"Delta"}
	for _, command := range commands {
		fmt.Println(command)
	}

	var client = lspaxos.Client{ClientId: 2048, Replicas: replicas[:]}
	client.StartClient(commands)*/
=======
	lspaxos.StartAcceptorServer(0, "1234")
>>>>>>> 7dca20c3851bdedbd9c2463ce3c222ae2afa9b45
}
