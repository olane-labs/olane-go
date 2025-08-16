package config

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

func TestDefaultLibp2pConfig(t *testing.T) {
	config := DefaultLibp2pConfig()

	// Test default values
	if len(config.Listeners) != 1 || config.Listeners[0] != "/ip4/0.0.0.0/tcp/0" {
		t.Errorf("Expected default listener '/ip4/0.0.0.0/tcp/0', got %v", config.Listeners)
	}

	if config.Identity == nil {
		t.Error("Expected identity to be generated")
	}

	if config.ConnMgr == nil {
		t.Error("Expected connection manager to be created")
	}

	if !config.EnableRelay {
		t.Error("Expected relay to be enabled by default")
	}

	if !config.EnableDHT {
		t.Error("Expected DHT to be enabled by default")
	}

	if !config.EnablePubsub {
		t.Error("Expected pubsub to be enabled by default")
	}

	if config.KBucketSize != 20 {
		t.Errorf("Expected k-bucket size to be 20, got %d", config.KBucketSize)
	}
}

func TestCreateNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with default config
	h, dht, pubsub, err := CreateNode(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to create node with default config: %v", err)
	}
	defer h.Close()
	defer func() {
		if dht != nil {
			dht.Close()
		}
	}()

	if h == nil {
		t.Error("Expected host to be created")
	}

	if dht == nil {
		t.Error("Expected DHT to be created")
	}

	if pubsub == nil {
		t.Error("Expected pubsub to be created")
	}

	// Test with custom config
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, nil)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	connMgr, err := connmgr.NewConnManager(50, 200, connmgr.WithGracePeriod(30))
	if err != nil {
		t.Fatalf("Failed to create connection manager: %v", err)
	}

	customConfig := &Libp2pConfig{
		Listeners:      []string{"/ip4/127.0.0.1/tcp/0"},
		Identity:       priv,
		ConnMgr:        connMgr,
		EnableRelay:    false,
		EnableDHT:      true,
		EnablePubsub:   true,
		KBucketSize:    10,
		BootstrapPeers: []string{},
	}

	h2, dht2, pubsub2, err := CreateNode(ctx, customConfig)
	if err != nil {
		t.Fatalf("Failed to create node with custom config: %v", err)
	}
	defer h2.Close()
	defer func() {
		if dht2 != nil {
			dht2.Close()
		}
	}()

	if h2 == nil {
		t.Error("Expected host to be created with custom config")
	}

	if dht2 == nil {
		t.Error("Expected DHT to be created with custom config")
	}

	if pubsub2 == nil {
		t.Error("Expected pubsub to be created with custom config")
	}

	// Verify the nodes have different IDs
	if h.ID() == h2.ID() {
		t.Error("Expected different peer IDs for different nodes")
	}
}

func TestCreateNodeDisabledServices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := DefaultLibp2pConfig()
	config.EnableDHT = false
	config.EnablePubsub = false

	h, dht, pubsub, err := CreateNode(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create node with disabled services: %v", err)
	}
	defer h.Close()

	if h == nil {
		t.Error("Expected host to be created")
	}

	if dht != nil {
		t.Error("Expected DHT to be nil when disabled")
	}

	if pubsub != nil {
		t.Error("Expected pubsub to be nil when disabled")
	}
}

func TestConnectToBootstrapPeers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h, dht, _, err := CreateNode(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer h.Close()
	defer func() {
		if dht != nil {
			dht.Close()
		}
	}()

	// Test with empty bootstrap peers
	err = ConnectToBootstrapPeers(ctx, h, []string{})
	if err != nil {
		t.Errorf("Expected no error with empty bootstrap peers, got: %v", err)
	}

	// Test with invalid bootstrap peer
	err = ConnectToBootstrapPeers(ctx, h, []string{"invalid-multiaddr"})
	if err == nil {
		t.Error("Expected error with invalid bootstrap peer")
	}
}
