// Package node provides utilities for creating and managing Olane network nodes.
// This package mirrors the functionality of the TypeScript o-config/node package.
package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	"github.com/olane-labs/olane-go/pkg/config"
)

// Node represents an Olane network node with libp2p capabilities
type Node struct {
	Host       host.Host
	DHT        *dht.IpfsDHT
	PubSub     *pubsub.PubSub
	Config     *config.Libp2pConfig
	ctx        context.Context
	cancelFunc context.CancelFunc
	mu         sync.RWMutex
	isRunning  bool
}

// NewNode creates a new Olane network node with the given configuration
func NewNode(ctx context.Context, cfg *config.Libp2pConfig) (*Node, error) {
	if cfg == nil {
		cfg = config.DefaultLibp2pConfig()
	}

	nodeCtx, cancel := context.WithCancel(ctx)

	// Create the libp2p host and services
	h, kadDHT, gossipSub, err := config.CreateNode(nodeCtx, cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	node := &Node{
		Host:       h,
		DHT:        kadDHT,
		PubSub:     gossipSub,
		Config:     cfg,
		ctx:        nodeCtx,
		cancelFunc: cancel,
		isRunning:  false,
	}

	return node, nil
}

// Start starts the node and connects to bootstrap peers
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isRunning {
		return fmt.Errorf("node is already running")
	}

	// Connect to bootstrap peers
	if err := config.ConnectToBootstrapPeers(n.ctx, n.Host, n.Config.BootstrapPeers); err != nil {
		return fmt.Errorf("failed to connect to bootstrap peers: %w", err)
	}

	n.isRunning = true
	return nil
}

// Stop gracefully shuts down the node
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isRunning {
		return nil
	}

	// Close services in reverse order
	if n.PubSub != nil {
		// PubSub doesn't have a Close method, it's cleaned up when the host closes
	}

	if n.DHT != nil {
		if err := n.DHT.Close(); err != nil {
			fmt.Printf("Warning: error closing DHT: %v\n", err)
		}
	}

	if n.Host != nil {
		if err := n.Host.Close(); err != nil {
			fmt.Printf("Warning: error closing host: %v\n", err)
		}
	}

	n.cancelFunc()
	n.isRunning = false
	return nil
}

// IsRunning returns whether the node is currently running
func (n *Node) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isRunning
}

// ID returns the peer ID of this node
func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

// Addrs returns the multiaddresses this node is listening on
func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.Host.Addrs()
}

// Peers returns the list of connected peers
func (n *Node) Peers() []peer.ID {
	return n.Host.Network().Peers()
}

// PeerCount returns the number of connected peers
func (n *Node) PeerCount() int {
	return len(n.Peers())
}

// ConnectToPeer connects to a specific peer
func (n *Node) ConnectToPeer(ctx context.Context, peerAddr string) error {
	ma, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address %s: %w", peerAddr, err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return fmt.Errorf("failed to parse peer info from %s: %w", peerAddr, err)
	}

	return n.Host.Connect(ctx, *peerInfo)
}

// Subscribe subscribes to a pubsub topic
func (n *Node) Subscribe(topic string) (*pubsub.Subscription, error) {
	if n.PubSub == nil {
		return nil, fmt.Errorf("pubsub is not enabled on this node")
	}

	return n.PubSub.Subscribe(topic)
}

// Publish publishes data to a pubsub topic
func (n *Node) Publish(ctx context.Context, topic string, data []byte) error {
	if n.PubSub == nil {
		return fmt.Errorf("pubsub is not enabled on this node")
	}

	topicHandle, err := n.PubSub.Join(topic)
	if err != nil {
		return fmt.Errorf("failed to join topic %s: %w", topic, err)
	}
	defer topicHandle.Close()

	return topicHandle.Publish(ctx, data)
}

// GetValue retrieves a value from the DHT
func (n *Node) GetValue(ctx context.Context, key string) ([]byte, error) {
	if n.DHT == nil {
		return nil, fmt.Errorf("DHT is not enabled on this node")
	}

	return n.DHT.GetValue(ctx, key)
}

// PutValue stores a value in the DHT
func (n *Node) PutValue(ctx context.Context, key string, value []byte) error {
	if n.DHT == nil {
		return fmt.Errorf("DHT is not enabled on this node")
	}

	return n.DHT.PutValue(ctx, key, value)
}

// FindPeer finds a peer in the DHT
func (n *Node) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	if n.DHT == nil {
		return peer.AddrInfo{}, fmt.Errorf("DHT is not enabled on this node")
	}

	return n.DHT.FindPeer(ctx, peerID)
}

// Context returns the node's context
func (n *Node) Context() context.Context {
	return n.ctx
}
