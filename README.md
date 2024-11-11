## Distributed mutex
Running the System
Starting Nodes
To run the system, each node in the distributed system must be started separately on different machines or the same machine with different ports. You can specify the node ID, the address of the next node in the token ring, and whether the node starts with the token.

Run the following command to start a node:

bash
Copiar código
go run node.go <nodeID> <nextNode> <initialToken>
<nodeID>: The unique identifier for this node. It should be an integer.
<nextNode>: The address of the next node in the ring (e.g., localhost:5002 for the next node).
<initialToken>: Whether the node should start with the token (true or false). Only the node that starts with the token can access the critical section initially.
Example:
Start Node 1 with the token:

bash
Copiar código
go run node.go 1 localhost:5002 true
Start Node 2 (without the token):

bash
Copiar código
go run node.go 2 localhost:5003 false
Start Node 3 (without the token):

bash
Copiar código
go run node.go 3 localhost:5001 false
