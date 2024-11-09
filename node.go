package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tcapook01/distributed_mutex/proto/nodepb" // Asegúrate de que esta ruta sea correcta

	"google.golang.org/grpc"
)

type Node struct {
	nodepb.UnimplementedNodeServer
	id           int32
	timestamp    int64
	requestingCS bool
	peers        []string
	grpcClients  map[string]nodepb.NodeClient
	mutex        sync.Mutex
	grantChannel chan bool
}

func NewNode(id int32, peers []string) *Node {
	return &Node{
		id:           id,
		timestamp:    0,
		requestingCS: false,
		peers:        peers,
		grpcClients:  make(map[string]nodepb.NodeClient),
		grantChannel: make(chan bool, 1),
	}
}

// RequestAccess RPC to handle incoming requests to enter the CS
func (n *Node) RequestAccess(ctx context.Context, req *nodepb.AccessRequest) (*nodepb.AccessReply, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.timestamp = max(n.timestamp, req.Timestamp) + 1
	log.Printf("Node %d received access request from Node %d with timestamp %d", n.id, req.NodeId, req.Timestamp)

	// Check conditions to grant access or delay
	if !n.requestingCS || (req.Timestamp < n.timestamp || (req.Timestamp == n.timestamp && req.NodeId < n.id)) {
		// Convert NodeId to string and use it as key to access grpcClients
		client, ok := n.grpcClients[strconv.Itoa(int(req.NodeId))]
		if ok {
			log.Printf("Node %d is granting access to Node %d", n.id, req.NodeId)
			client.GrantAccess(ctx, &nodepb.AccessGrant{NodeId: n.id})
		}
	}

	return &nodepb.AccessReply{}, nil
}

// GrantAccess RPC to receive grants from other nodes
func (n *Node) GrantAccess(ctx context.Context, grant *nodepb.AccessGrant) (*nodepb.Ack, error) {
	log.Printf("Node %d received grant from Node %d", n.id, grant.NodeId)

	// Signal that grant was received
	n.grantChannel <- true

	return &nodepb.Ack{}, nil
}

// Method to initiate request to enter CS
func (n *Node) requestCriticalSection() {
	n.mutex.Lock()
	n.timestamp++
	n.requestingCS = true
	n.mutex.Unlock()

	// Send request to all peers
	for _, addr := range n.peers {
		client := n.grpcClients[addr]
		go func(client nodepb.NodeClient) {
			client.RequestAccess(context.Background(), &nodepb.AccessRequest{NodeId: n.id, Timestamp: n.timestamp})
		}(client)
	}

	// Wait for grants from all peers
	for range n.peers {
		<-n.grantChannel
	}

	// Enter CS
	n.enterCriticalSection()
}

// Emulate entering the critical section
func (n *Node) enterCriticalSection() {
	log.Printf("Node %d is entering the Critical Section", n.id)
	time.Sleep(2 * time.Second) // Simulate critical section work
	log.Printf("Node %d is leaving the Critical Section", n.id)

	n.mutex.Lock()
	n.requestingCS = false
	n.mutex.Unlock()
}

// Helper function to get the max of two timestamps
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run node.go <node_id>")
	}

	// Convert nodeID from string to int
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid node ID: %v", err)
	}

	// Convert nodeID to int32 for use in Node struct
	nodeIDInt32 := int32(nodeID)

	var peers []string
	switch nodeID {
	case 1:
		peers = []string{"localhost:5002", "localhost:5003"}
	case 2:
		peers = []string{"localhost:5001", "localhost:5003"}
	case 3:
		peers = []string{"localhost:5001", "localhost:5002"}
	}

	node := NewNode(nodeIDInt32, peers)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", 5000+nodeID))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", 5000+nodeID, err)
	}

	grpcServer := grpc.NewServer()
	nodepb.RegisterNodeServer(grpcServer, node)
	log.Printf("Node %d listening on port %d", nodeID, 5000+nodeID)

	// Establish client connections to peers
	for _, addr := range peers {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to connect to peer at %s: %v", addr, err)
		}
		defer conn.Close()
		node.grpcClients[addr] = nodepb.NewNodeClient(conn)
	}

	// Start the gRPC server
	go grpcServer.Serve(listen)

	// Make nodes request access to the critical section periodically
	for {
		// Simulate random delay before making each request (example: 5-10 seconds)
		time.Sleep(time.Duration(5+node.id) * time.Second) // Random delay between requests
		node.requestCriticalSection()                      // Request access to CS
	}
}
