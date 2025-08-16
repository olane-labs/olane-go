// Package main provides an example of how to use the olane-go library.
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
	"github.com/olane-labs/olane-go/pkg/node"
)

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

	// Create a custom configuration
	cfg := config.DefaultLibp2pConfig()
	cfg.Listeners = []string{"/ip4/0.0.0.0/tcp/4001"}
	
	// Add some bootstrap peers (these are standard IPFS bootstrap nodes)
	cfg.BootstrapPeers = []string{
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.236.179.241/tcp/4001/p2p/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	}

	// Create and start the node
	n, err := node.NewNode(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	fmt.Printf("Node created with ID: %s\n", n.ID())
	fmt.Printf("Listening on addresses:\n")
	for _, addr := range n.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, n.ID())
	}

	if err := n.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	fmt.Println("Node started successfully!")

	// Example: Subscribe to a topic
	if n.PubSub != nil {
		topic := "olane-example"
		sub, err := n.Subscribe(topic)
		if err != nil {
			log.Printf("Failed to subscribe to topic %s: %v", topic, err)
		} else {
			fmt.Printf("Subscribed to topic: %s\n", topic)
			
			// Start a goroutine to handle incoming messages
			go func() {
				for {
					msg, err := sub.Next(ctx)
					if err != nil {
						if ctx.Err() != nil {
							// Context cancelled, exit gracefully
							return
						}
						log.Printf("Error reading message: %v", err)
						continue
					}
					fmt.Printf("Received message from %s: %s\n", msg.GetFrom(), string(msg.GetData()))
				}
			}()

			// Send a test message every 30 seconds
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						message := fmt.Sprintf("Hello from %s at %s", n.ID(), time.Now().Format(time.RFC3339))
						if err := n.Publish(ctx, topic, []byte(message)); err != nil {
							log.Printf("Failed to publish message: %v", err)
						} else {
							fmt.Printf("Published message: %s\n", message)
						}
					}
				}
			}()
		}
	}

	// Example: Store and retrieve a value in the DHT
	if n.DHT != nil {
		go func() {
			// Wait a bit for the DHT to be ready
			time.Sleep(10 * time.Second)
			
			key := "olane-example-key"
			value := []byte("Hello from Olane DHT!")
			
			fmt.Printf("Storing value in DHT with key: %s\n", key)
			if err := n.PutValue(ctx, key, value); err != nil {
				log.Printf("Failed to put value in DHT: %v", err)
			} else {
				fmt.Println("Value stored successfully in DHT")
				
				// Try to retrieve it
				time.Sleep(2 * time.Second)
				fmt.Printf("Retrieving value from DHT with key: %s\n", key)
				retrievedValue, err := n.GetValue(ctx, key)
				if err != nil {
					log.Printf("Failed to get value from DHT: %v", err)
				} else {
					fmt.Printf("Retrieved value: %s\n", string(retrievedValue))
				}
			}
		}()
	}

	// Print peer information periodically
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				peerCount := n.PeerCount()
				fmt.Printf("Connected to %d peers\n", peerCount)
				if peerCount > 0 {
					fmt.Println("Peer IDs:")
					for _, peerID := range n.Peers() {
						fmt.Printf("  %s\n", peerID)
					}
				}
			}
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Stop the node
	fmt.Println("Stopping node...")
	if err := n.Stop(); err != nil {
		log.Printf("Error stopping node: %v", err)
	} else {
		fmt.Println("Node stopped successfully")
	}
}
