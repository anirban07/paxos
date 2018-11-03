package main

import (
	"lspaxos"
)

func main() {
	lspaxos.StartAcceptorServer(0, "1234")
}
