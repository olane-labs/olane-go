// Package main provides an example of how to use the olane-go core functionality.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olane-labs/olane-go/pkg/config"
	"github.com/olane-labs/olane-go/pkg/core"
)

// ExampleNode demonstrates a concrete implementation of a CoreNode
type ExampleNode struct {
	*core.CoreNode
}

// NewExampleNode creates a new example node
func NewExampleNode(cfg *core.CoreConfig) *ExampleNode {
	coreNode := core.NewCoreNode(cfg)
	return &ExampleNode{
		CoreNode: coreNode,
	}
}

// Initialize implements the node initialization
func (n *ExampleNode) Initialize(ctx context.Context) error {
	fmt.Printf("Initializing example node: %s\n", n.Address().String())
	
	// For now, we'll just call the parent Initialize method
	// In a real implementation, this would create the libp2p node and set up services
	err := n.CoreNode.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize core node: %w", err)
	}

	fmt.Printf("Example node initialized successfully\n")
	return nil
}

func main() {
	// Create a context that will be cancelled on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived shutdown signal, stopping node...")
		cancel()
	}()

	// Create core configuration
	cfg := core.DefaultCoreConfig()
	cfg.Address = core.NewOAddress("o://example-node")
	cfg.Type = core.NodeTypeNode
	cfg.Name = "example"
	cfg.Description = "An example Olane node implemented in Go"
	
	// Configure network settings
	networkCfg := config.DefaultLibp2pConfig()
	networkCfg.Listeners = []string{"/ip4/0.0.0.0/tcp/4001"}
	cfg.Network = networkCfg

	// Add some example methods
	cfg.Methods = map[string]*core.OMethod{
		"hello": {
			Name:        "hello",
			Description: "Returns a greeting message",
			Parameters: map[string]interface{}{
				"name": "string",
			},
			Returns: map[string]interface{}{
				"message": "string",
			},
		},
		"info": {
			Name:        "info",
			Description: "Returns node information",
			Parameters:  map[string]interface{}{},
			Returns: map[string]interface{}{
				"nodeInfo": "object",
			},
		},
	}

	// Create and start the node
	node := NewExampleNode(cfg)

	fmt.Printf("Starting node: %s\n", node.Address().String())
	fmt.Printf("Node type: %s\n", node.Type())

	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	fmt.Println("Node started successfully!")

	// Demonstrate address functionality
	fmt.Println("\n=== Address Examples ===")
	
	// Create some example addresses
	addresses := []*core.OAddress{
		core.NewOAddress("o://leader"),
		core.NewOAddress("o://tools/calculator"),
		core.NewOAddress("o://services/weather/current"),
		core.NewOAddress("o://ai/gpt/chat"),
	}

	for _, addr := range addresses {
		fmt.Printf("Address: %s\n", addr.String())
		fmt.Printf("  Root: %s\n", addr.Root())
		fmt.Printf("  Paths: %s\n", addr.Paths())
		fmt.Printf("  Protocol: %s\n", addr.Protocol())
		fmt.Printf("  Is Leader: %t\n", addr.IsLeaderAddress())
		fmt.Printf("  Is Tool: %t\n", addr.IsToolAddress())
		if addr.IsToolAddress() {
			fmt.Printf("  Tool Name: %s\n", addr.GetToolName())
		}
		fmt.Printf("  Method: %s\n", addr.GetMethod())
		
		// Generate CID for the address
		if cid, err := addr.ToCID(); err == nil {
			fmt.Printf("  CID: %s\n", cid.String())
		}
		fmt.Println()
	}

	// Demonstrate node information
	fmt.Println("=== Node Information ===")
	if whoami, err := node.WhoAmI(ctx); err == nil {
		fmt.Printf("Address: %s\n", whoami.Address)
		fmt.Printf("Type: %s\n", whoami.Type)
		fmt.Printf("Description: %s\n", whoami.Description)
		fmt.Printf("Success Count: %d\n", whoami.SuccessCount)
		fmt.Printf("Error Count: %d\n", whoami.ErrorCount)
		fmt.Printf("Methods: %d available\n", len(whoami.Methods))
		for name, method := range whoami.Methods {
			fmt.Printf("  - %s: %s\n", name, method.Description)
		}
	}

	// Demonstrate child address creation
	fmt.Println("\n=== Child Address Examples ===")
	parentAddr := core.NewOAddress("o://network")
	childAddr := core.NewOAddress("o://service")
	combinedAddr := core.ChildAddress(parentAddr, childAddr)
	fmt.Printf("Parent: %s + Child: %s = Combined: %s\n", 
		parentAddr.String(), childAddr.String(), combinedAddr.String())

	// Demonstrate address validation
	fmt.Println("\n=== Address Validation ===")
	validAddr := core.NewOAddress("o://valid/address")
	invalidAddr := core.NewOAddress("invalid://address")
	fmt.Printf("Valid address '%s': %t\n", validAddr.String(), validAddr.Validate())
	fmt.Printf("Invalid address '%s': %t\n", invalidAddr.String(), invalidAddr.Validate())

	// Monitor node state
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Printf("Node state: %s, Errors: %d\n", node.State(), len(node.Errors()))
			}
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Stop the node
	fmt.Println("Stopping node...")
	if err := node.Stop(ctx); err != nil {
		log.Printf("Error stopping node: %v", err)
	} else {
		fmt.Println("Node stopped successfully")
	}
}
