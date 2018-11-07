# Paxos Made From Scratch
#### Authors:
Anirban Biswas and Robby Marver

## Assumptions Made
In this assignment we made our assumptions based on the problem specification. Specifically: 

1. Once a node fails, it will never recover
2. Nodes do not resend messages
3. A majority of messages will always be successfully sent and received
4. Message drops are indistinguishable from node failures in an asynchronous environment (aka The network is perfect).

From assumption 2, assumption 3 follows logically. If a node never resends a message, then progress can only be achieved if at least a majority of messages are sent and received. This still allows for failures, but only less than a majority of them.

Assumption 4 is purely for convenience of testing.  We found it was easier to kill a node than it was to interrupt the message. In an asynchronous environment, because all you know was that the message was enqueued on the network. If you don't get a response, it could be because the network dropped the sent message, or the node died, or the response was dropped. Because we kill nodes instead of dropping messages, the implicit assumption here is that the network is perfect.

## Detailed Implementation Description
We chose to implement this project using Golang. The implementation of the roles were based on roles defined in Paxos Made Moderately Complex. All roles defined below maintain a unique identifier within their role (i.e. You can have Leader 0 and Replica 0, but you cannot have two Leader 0's).
### Infrastructure
* We used Golang's RPC library as well as Goroutines and channels to send/receive asynchronously between the different roles.  Specifically, a role would send a message by spawning a go-routine to send a message using an RPC, which would write a response to a channel on the role that spawned the message.  Because of this, our message handling is done on a send-receive pattern instead of an event-handler pattern.  The "events" in this case are the RPC returns, where they are handled when the role is waiting to read from a channel.
* We interpret commands to be unique only on the client ID that sent the command and the sequence number of that client. A client may then send two lock requests on the same lock in succession and they will be interpreted differently if the sequence numbers are different. The implication here is that the client can issue logically-duplicate requests.
* We communicate over TCP ports so you could theoretically run our solution on different machines and it would still work (as long as the addresses were correct).
* Our test suite is written in lspaxos/test_test.go. Here we test different configurations of the roles (single clients, multiple replicas, multiple leaders, etc.) as well as failure cases (leader failures, replica failures, acceptor failures).
* All of the message/command types are defined in lspaxos/common.go
* We also provide the appropriate functions to create a client-server interaction of Paxos as an example of what the LockServer application might look like with clients. This is located in main/main.go

### Client
The client maintains the following state:

- A msgID that keeps track of which message the client is expecting a response to.
- The addresses of the replicas.
- A timeout that maintains how long the client should sleep if it requested a lock that is held by another client.

The client will sequentially execute commands. It will not advance to the next command until the current command has been responded to with a valid response. The client can issue two types of commands: Lock and Unlock. For lock, the possible responses are:

- OK: The lock the client requested is now held by this client (i.e. it was not previously owned by anyone or the client already held the lock).
- ErrLockHeld: The lock the client requested is held by another client. The client will sleep using AIMD and then retry the lock request. It will increment its message ID.

For unlock, the possible responses are:

- OK: The lock was owned by this client and is now free.
- ErrInvalidUnlock: The lock was owned by another client or was free. This is a no-op for the client, so it should just continue as if nothing happened.

See main.go for an example of how to use the client with the Paxos servers. You can read in a client specification using ReadSpec(). See specs/test_spec.txt for an example of what the specifications should look like.

### Replica
The replica maintains the following state:

- A map of lock names to their client ID owners (the lock server state)
- A slot number (slotIn) indicating the next slot that is not associated with any request in the log
- A slot number (slotOut) indicating the next slot that has not been decided yet in the log.
- A set of requests that have been received from clients
- A map of slots to commands that keep track of which requests are currently being decided by Paxos
- A map of slots to commands that keep track of what commands have already been decided by Paxos
- A set of addresses of the leaders
- A socket listener to accept incoming client requests
- The address of the replica to set up the socket listener
- A debug variable dead, which is a flag to indicate if this replica has died (used in tests)

The replica has the following things running simultaneously:

1. propose(): propose will wait for incoming client requests. Once it has been notified that there are requests from clients, it will move those requests into its proposals map and start a round of Paxos for each command. It will increment slotIn for each round that it starts.
2. perform(): perform will wait for responses from leaders of Paxos. Once it receives a decision from the leaders on a particular slot, it will update its decisions map. It will then try to perform, in order, all commands starting from slotOut in the decisions map. Decisions that are out of order or that are not sequential will not be performed on the replica's state (i.e. decision 2 will not be performed until slot 1 has been decided). Importantly, if the command the replica proposed for the decided slot is not the same as the command that was ultimately decided for that slot, that command will be moved back into the requests set and perform() will notify propose() to start a new round.
3. ExecuteRequest: Upon receiving a request from a client, the command associated with that request will be added to the requests map, and propose() will be notified. Then ExecuteRequest will block until perform() notifies it that something has been decided.  Upon receiving the notification, ExecuteRequest will create a local lockMap and replay the log until it finds that the command from the client is in the log. It will replay the log in order and make sure the at all locks/unlocks are valid.  It must replay the log to ensure correctness because the Leader/Acceptor has no notion of correctness and perform() may have performed several requests in a row before notifying ExecuteRequest that something was performed (i.e. how do you know if an unlock was valid?).  It will then respond to the client if the command the client requested was decided. 


### Leader
The leader maintains the following state:

- A ballot number that uniquely identifies the leader. Ballots are defined as a Number, and a Leader, where the Number takes precedence when comparing ballots (i.e. 1.1 > 0.1, 1.1 > 1.0)
- The addresses of all acceptors
- A flag maintaining if the leader is a commander
- An integer indicating the amount of responses from acceptors needed to reach a majority
- A timeout maintaining how long the leader should wait before trying to become the commander
- A map of slots to commands keeping track of what is currently being proposed to acceptors
- A map of slots to commands keeping track of what has been decided
- A listening socket to accept incoming RPC calls
- An address, used to initialize the listening socket
- A debug variable dead, which is a flag to indicate if this leader has died (used in tests)

The leader essentially has three things happening simultaneously:

1. A scout() thread is always running, trying to elect this leader to become a commander. An election is when a majority of acceptors have accepted this leader's ballot number as the highest ballot number they have seen. During the election process, the leader will learn of values that it should propose for certain slots that the acceptors it has contacted have already accepted. This maintains the invariant that the leaders will only propose values that have been previously accepted, if any. If the scout thread receives a ballot number that is higher than its own ballot, it means that some other leader is the commander, and so this leader will sleep using AIMD before increasing it's ballot number and trying again to become commander.
2. Once a leader becomes the commander, it will spawn a commander thread for every slot number in its proposals map. This is how the leader learns what has been previously decided if it was not previously the commander. Commander threads are slot independent, in that two commanders can be simultaneously trying to have a majority of acceptors accept their command on different slots (i.e. for every slot, there is one commander).
3. The leader is also listening for requests from a replica. On a replica request, the leader will add the proposed command from the replica to its proposals map if that proposals map did not already have a command associated with the slot the replica was proposing on. It will also check its decisions map to make sure the replica is not trying to retry a slot that was decided.  If the leader is the commander, it will spawn a commander thread for this slot. No matter what, this ExecutePropose call will block until it is notified by a commander that a decision has been made, and it will then and only then tell the replica that this slot has been decided. The command on the slot may not be the same as the command that the replica proposed to the leader.


### Acceptor
The acceptor maintains the following state:

- A ballot number that is the highest ballot number this acceptor has accepted.
- A map of slot number to command that keeps track of the commands this acceptor has accepted.
- A listener, which is the open socket that is used to accept incoming RPC calls
- An address, used to initialize the listening socket
- A debug variable dead, which is a flag to indicate if this acceptor has died (used in tests)

The acceptor supports two message handlers:
##### 1. Execute Propose:
The acceptor receives a ballot number from a leader. If the ballot number is higher than the one it has previously accepted, the acceptor will update its ballot number to be the one it received from the leader. The acceptor will always respond with its ballot number, as well as its map of accepted commands.
##### 2. Execute Accept:
The acceptor receives a ballot number, a slot number, and a command from a leader. If the ballot number received is the same or greater than the highest accepted ballot number this acceptor has, the acceptor will update its ballot number, and accept the command for the slot given by the leader. It will then respond to the leader with its ballot number.


## How to use
Since each role in the protocol is an independent process, they need to be started individually. The roles of Replica, Leader and Acceptor each have `Start<Rolename>` methods. These methods take in unique identifiers, an address to listen on, and if needed, a list of addresses of servers they need to send RPCs to. The methods set up a listener on the given address. They also return structs that hold relevant state of the respective roles. These structs can be used to kill the server.

A client can be created using the `StartClient`. It returns a struct with the client's initial state. This struct is used to send lock and unlock requests defined in `Client.go`.

## Outstanding issues
There are no known outstanding issues according to the spec. However here are a few things that could be improved:

- The leaders do not communicate with each other. This means, a lagging leader must propose values for every missing slot number to learn the corresponding command.
- Unreliable networks can be handled using appropriate timeouts on the RPCs and resending requests.


## Anything else
When running all the tests in `test_test.go`, you may see unexpected closed connection error. These are messages from previous tests that were in flight, but arrived to their destinations after the destination was killed.