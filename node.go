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

	pb "github.com/tcapook01/distributed_mutex/proto" // replace with actual path

	"google.golang.org/grpc"
)

type TokenRingServer struct {
	pb.UnimplementedTokenRingServer
	nodeID   int32
	hasToken bool
	nextNode string
	mutex    sync.Mutex
}

// NewTokenRingServer creates a new server instance
func NewTokenRingServer(nodeID int32, nextNode string, initialToken bool) *TokenRingServer {
	return &TokenRingServer{
		nodeID:   nodeID,
		hasToken: initialToken,
		nextNode: nextNode,
	}
}

// PassToken receives the token from the previous node in the ring
func (s *TokenRingServer) PassToken(ctx context.Context, tokenMsg *pb.TokenMessage) (*pb.Response, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Print the TokenHolderId and the current node's ID to debug the mismatch
	fmt.Printf("Node %d received token message: TokenHolderId = %d, NodeID = %d\n", s.nodeID, tokenMsg.TokenHolderId, s.nodeID)

	// Check if the token holder ID matches the previous node's ID
	if tokenMsg.TokenHolderId == s.nodeID-1 || (s.nodeID == 1 && tokenMsg.TokenHolderId == 3) { // Adjust for first node case
		s.hasToken = true
		fmt.Printf("Node %d has received the token and is entering the critical section\n", s.nodeID)
		s.enterCriticalSection()
	} else {
		fmt.Printf("Node %d received token message but it does not match its ID. TokenHolderId = %d\n", s.nodeID, tokenMsg.TokenHolderId)
	}

	return &pb.Response{Message: "Token passed"}, nil
}

// enterCriticalSection simulates work in the critical section and then passes the token
func (s *TokenRingServer) enterCriticalSection() {
	fmt.Printf("Node %d is entering critical section\n", s.nodeID)
	time.Sleep(2 * time.Second) // simulate critical section work
	fmt.Printf("Node %d leaving critical section\n", s.nodeID)

	// After finishing the critical section, pass the token to the next node
	s.hasToken = false
	s.passToken() // Pass the token to the next node
}

// passToken sends the token to the next node in the ring
func (s *TokenRingServer) passToken() {
	s.hasToken = false
	for {
		conn, err := grpc.Dial(s.nextNode, grpc.WithInsecure())
		if err != nil {
			log.Printf("Node %d failed to connect to next node %s: %v", s.nodeID, s.nextNode, err)
			time.Sleep(1 * time.Second) // Retry after a delay
			continue
		}
		defer conn.Close()

		client := pb.NewTokenRingClient(conn)
		// Here we pass the current node's ID as TokenHolderId
		_, err = client.PassToken(context.Background(), &pb.TokenMessage{TokenHolderId: s.nodeID})
		if err != nil {
			log.Printf("Node %d failed to pass token: %v", s.nodeID, err)
			time.Sleep(1 * time.Second) // Retry after a delay
			continue
		}

		// Print that the node passed the token to the next node
		fmt.Printf("Node %d passed the token to node at %s\n", s.nodeID, s.nextNode)
		break
	}
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run node.go <nodeID> <nextNode> <initialToken>")
	}

	// Parse the nodeID as an int32
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid nodeID: %v", err)
	}
	nextNode := os.Args[2]
	initialToken := os.Args[3] == "true"

	// Create a new server instance
	server := NewTokenRingServer(int32(nodeID), nextNode, initialToken)

	// If the node is the initial holder, start the process by passing the token
	if initialToken {
		go func() {
			time.Sleep(1 * time.Second) // Wait a moment before starting the token pass
			server.passToken()
		}()
	}

	// Start the gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 5000+nodeID))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTokenRingServer(grpcServer, server)

	fmt.Printf("Node %d is starting\n", server.nodeID)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
