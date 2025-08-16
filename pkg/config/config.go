// Package config provides libp2p configuration utilities for Olane networks.
// This package mirrors the functionality of the TypeScript @olane/o-config package.
package config

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

// Libp2pConfig holds configuration options for libp2p nodes
type Libp2pConfig struct {
	// Listeners specifies the multiaddrs to listen on
	Listeners []string
	// BootstrapPeers is a list of bootstrap peer addresses
	BootstrapPeers []string
	// Identity holds the private key for the node
	Identity crypto.PrivKey
	// ConnMgr configures the connection manager
	ConnMgr *connmgr.BasicConnMgr
	// EnableRelay enables circuit relay functionality
	EnableRelay bool
	// EnableDHT enables the Kademlia DHT
	EnableDHT bool
	// EnablePubsub enables gossipsub
	EnablePubsub bool
	// DHTProtocolPrefix sets the DHT protocol prefix
	DHTProtocolPrefix protocol.ID
	// KBucketSize sets the DHT k-bucket size
	KBucketSize int
}

// DefaultLibp2pConfig returns a default configuration for libp2p nodes
// This mirrors the defaultLibp2pConfig from the TypeScript version
func DefaultLibp2pConfig() *Libp2pConfig {
	// Generate a new identity
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("failed to generate key pair: %v", err))
	}

	// Create a basic connection manager
	connMgr, err := connmgr.NewConnManager(
		100, // Low watermark
		400, // High watermark
		connmgr.WithGracePeriod(60), // Grace period in seconds
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create connection manager: %v", err))
	}

	return &Libp2pConfig{
		Listeners:         []string{"/ip4/0.0.0.0/tcp/0"},
		BootstrapPeers:    []string{},
		Identity:          priv,
		ConnMgr:           connMgr,
		EnableRelay:       true,
		EnableDHT:         true,
		EnablePubsub:      true,
		DHTProtocolPrefix: "/ipfs/kad/1.0.0",
		KBucketSize:       20,
	}
}

// CreateNode creates a libp2p node with the given configuration
// This mirrors the createNode function from the TypeScript version
func CreateNode(ctx context.Context, config *Libp2pConfig) (host.Host, *dht.IpfsDHT, *pubsub.PubSub, error) {
	if config == nil {
		config = DefaultLibp2pConfig()
	}

	// Convert listener strings to multiaddrs
	var listenAddrs []multiaddr.Multiaddr
	for _, addr := range config.Listeners {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid listen address %s: %w", addr, err)
		}
		listenAddrs = append(listenAddrs, ma)
	}

	// Build libp2p options
	opts := []libp2p.Option{
		// Identity
		libp2p.Identity(config.Identity),
		// Listen addresses
		libp2p.ListenAddrs(listenAddrs...),
		// Transports
		libp2p.Transport(tcp.NewTCPTransport),
		// Security
		libp2p.Security(noise.ID, noise.New),
		// Stream multiplexer
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		// Connection manager
		libp2p.ConnectionManager(config.ConnMgr),
		// Enable NAT traversal
		libp2p.NATPortMap(),
		// Enable AutoRelay if configured
	}

	if config.EnableRelay {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}))
	}

	// Create the libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	var kademliaDHT *dht.IpfsDHT
	var gossipSub *pubsub.PubSub

	// Initialize DHT if enabled
	if config.EnableDHT {
		kademliaDHT, err = dht.New(ctx, h,
			dht.Mode(dht.ModeServer),
			dht.ProtocolPrefix(config.DHTProtocolPrefix),
			dht.BucketSize(config.KBucketSize),
		)
		if err != nil {
			h.Close()
			return nil, nil, nil, fmt.Errorf("failed to create DHT: %w", err)
		}

		// Bootstrap the DHT
		if err = kademliaDHT.Bootstrap(ctx); err != nil {
			h.Close()
			kademliaDHT.Close()
			return nil, nil, nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
		}
	}

	// Initialize PubSub if enabled
	if config.EnablePubsub {
		gossipSub, err = pubsub.NewGossipSub(ctx, h,
			pubsub.WithMessageSigning(true),
			pubsub.WithStrictSignatureVerification(true),
		)
		if err != nil {
			h.Close()
			if kademliaDHT != nil {
				kademliaDHT.Close()
			}
			return nil, nil, nil, fmt.Errorf("failed to create pubsub: %w", err)
		}
	}

	return h, kademliaDHT, gossipSub, nil
}

// ConnectToBootstrapPeers connects the host to bootstrap peers
func ConnectToBootstrapPeers(ctx context.Context, h host.Host, bootstrapPeers []string) error {
	if len(bootstrapPeers) == 0 {
		return nil
	}

	for _, peerAddr := range bootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			return fmt.Errorf("invalid bootstrap peer address %s: %w", peerAddr, err)
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			return fmt.Errorf("failed to parse peer info from %s: %w", peerAddr, err)
		}

		if err := h.Connect(ctx, *peerInfo); err != nil {
			// Log warning but don't fail - bootstrap connection might be temporary
			fmt.Printf("Warning: failed to connect to bootstrap peer %s: %v\n", peerAddr, err)
		}
	}

	return nil
}
