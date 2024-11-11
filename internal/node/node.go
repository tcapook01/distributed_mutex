package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/tcapook01/distributed_mutex/proto/nodepb" // Asegúrate de que esta ruta sea correcta

	"google.golang.org/grpc"
)

// Represents a single node in the distributed system
type Node struct {
	nodepb.UnimplementedNodeServer
	id              int32
	timestamp       int64
	requestingCS    bool
	peers           map[int32]string // peerId -> address
	grpcClients     map[int32]nodepb.NodeClient
	connections     map[int32]*grpc.ClientConn // peerID -> connection
	mutex           sync.Mutex
	grantChannel    chan bool
	deferredRequest []nodepb.AccessRequest
}

// Initializes a new Node instance
func NewNode(id int32, peers map[int32]string) *Node {
	return &Node{
		id:              id,
		timestamp:       0,
		requestingCS:    false,
		peers:           peers,
		grpcClients:     make(map[int32]nodepb.NodeClient),
		connections:     make(map[int32]*grpc.ClientConn),
		grantChannel:    make(chan bool, len(peers)),
		deferredRequest: []nodepb.AccessRequest{},
	}
}

// RequestAccess RPC to handle incoming requests to enter the CS
func (n *Node) RequestAccess(ctx context.Context, req *nodepb.AccessRequest) (*nodepb.AccessReply, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Update logical timestamp
	if req.Timestamp > n.timestamp {
		n.timestamp = req.Timestamp
	}
	n.timestamp++

	log.Printf("Node %d received access request from Node %d with timestamp %d", n.id, req.NodeId, req.Timestamp)

	// Check conditions to grant access or delay
	shouldGrant := false
	if !n.requestingCS {
		shouldGrant = true
	} else {
		if req.Timestamp < n.timestamp {
			shouldGrant = true
		} else if req.Timestamp == n.timestamp && req.NodeId < n.id {
			shouldGrant = true
		}
	}

	if shouldGrant {
		// Grant access by sending GrantAccess RPC
		client, exists := n.grpcClients[req.NodeId]
		if exists {
			go func(client nodepb.NodeClient) {
				_, err := client.GrantAccess(context.Background(), &nodepb.AccessGrant{
					NodeId:    n.id,
					Timestamp: n.timestamp,
				})
				if err != nil {
					log.Printf("Failed to send GrantAccess to Node %d: %v", req.NodeId, err)
				} else {
					log.Printf("Node %d granted access to Node %d", n.id, req.NodeId)
				}
			}(client)
		} else {
			log.Printf("No client found for Node %d", req.NodeId)
		}
	} else {
		// Defer the reply
		n.deferredRequest = append(n.deferredRequest, *req)
		log.Printf("Node %d deferred access request from Node %d", n.id, req.NodeId)
	}

	// no reply payload needed
	return &nodepb.AccessReply{}, nil
}

// Handles incoming grant messages from other nodes
func (n *Node) GrantAccess(ctx context.Context, grant *nodepb.AccessGrant) (*nodepb.Ack, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

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
	currentTimestamp := n.timestamp
	n.mutex.Unlock()

	log.Printf("Node %d is requesting access to the critical section with timestamp %d", n.id, currentTimestamp)

	// Send request to all peers
	for peerID := range n.peers {
		go n.sendRequestAccess(peerID, currentTimestamp)
	}

	// Wait for grants from all peers
	grantReceived := 0
	for grantReceived < len(n.peers) {
		<-n.grantChannel
		grantReceived++
	}

	// Enter CS
	n.enterCriticalSection()
}

// sends a RequestAccess RPC to a specific peer
func (n *Node) sendRequestAccess(peerID int32, timestamp int64) {
	n.mutex.Lock()
	client, exists := n.grpcClients[peerID]
	n.mutex.Unlock()

	if !exists {
		log.Printf("No client found for Node %d. Skipping RequestAccess.", peerID)
		return
	}

	req := &nodepb.AccessRequest{
		NodeId:    n.id,
		Timestamp: timestamp,
	}

	maxRetries := 5
	backoff := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := client.RequestAccess(context.Background(), req)
		if err != nil {
			log.Printf("Attempt %d: Failed to send RequestAccess to Node %d: %v", attempt, peerID, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		log.Printf("Node %d successfully sent RequestAccess to Node %d", n.id, peerID)
		return
	}

	log.Printf("Failed to send RequestAccess to Node %d after %d attempts", peerID, maxRetries)
}

// Emulate entering the critical section
func (n *Node) enterCriticalSection() {
	log.Printf("Node %d is entering the Critical Section", n.id)
	fmt.Printf("Node %d is in the Critical Section\n", n.id)

	time.Sleep(2 * time.Second) // Simulate critical section work
	log.Printf("Node %d is leaving the Critical Section", n.id)

	n.mutex.Lock()
	n.requestingCS = false

	// Process deferred request
	for _, req := range n.deferredRequest {
		client, exists := n.grpcClients[req.NodeId]
		if exists {
			go func(client nodepb.NodeClient, req nodepb.AccessRequest) {
				_, err := client.GrantAccess(context.Background(), &nodepb.AccessGrant{
					NodeId:    n.id,
					Timestamp: n.timestamp,
				})
				if err != nil {
					log.Printf("Failed to send deferred GrantAccess to Node %d: %v", req.NodeId, err)
				} else {
					log.Printf("Node %d sent deferred GrantAccess to Node %d", n.id, req.NodeId)
				}
			}(client, req)
		}
	}
	// Clear the deferred request
	n.deferredRequest = []nodepb.AccessRequest{}
	n.mutex.Unlock()
}

func (n *Node) connectToPeers() {
	for peerID, addr := range n.peers {
		go func(peerID int32, addr string) {
			var conn *grpc.ClientConn
			var err error
			for {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					log.Printf("Failed to connect to peer %d at %s: %v. Retrying...", peerID, addr, err)
					time.Sleep(2 * time.Second)
					continue
				}
				log.Printf("Successfully connected to peer %d at %s", peerID, addr)
				n.mutex.Lock()
				n.grpcClients[peerID] = nodepb.NewNodeClient(conn)
				n.mutex.Unlock()
				break
			}
		}(peerID, addr)
	}
}

// Closes all gRPC client connection in a graceful way
func (n *Node) closeConnections() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for peerID, conn := range n.connections {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				log.Printf("Error closing connection to Node %d: %v", peerID, err)
			} else {
				log.Printf("Closed connection to Node %d", peerID)
			}
		}
	}
}

/*// Helper function to get the max of two timestamps
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}*/

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

	var peers map[int32]string
	switch nodeID {
	case 1:
		peers = map[int32]string{
			2: "localhost:5002",
			3: "localhost:5003"}
	case 2:
		peers = map[int32]string{
			1: "localhost:5001",
			3: "localhost:5003"}
	case 3:
		peers = map[int32]string{
			1: "localhost:5001",
			2: "localhost:5002"}
	default:
		log.Fatalf("Unsupported node ID: %d", nodeID)
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
	node.connectToPeers()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %s, shutting down gracefully...", sig)
		grpcServer.GracefulStop()
		node.closeConnections()
		os.Exit(0)
	}()

	// Start the gRPC server in a seperate goroutin
	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Seed the random nubmer generator
	rand.Seed(time.Now().UnixNano())

	// Make nodes request access to the critical section periodically
	for {
		// Simulate random delay before making each request (example: 5-10 seconds)
		sleepDuration := time.Duration(rand.Intn(6)+5) * time.Second
		time.Sleep(sleepDuration)
		node.requestCriticalSection() // Request access to CS
	}
}
