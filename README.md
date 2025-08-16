# Olane Go

A Go implementation of the Olane network infrastructure, providing peer-to-peer networking capabilities using libp2p.

## Overview

Olane-Go is a port of the TypeScript Olane project to Go, focusing on creating intelligent, self-improving network infrastructure. This implementation provides the core networking primitives needed to build decentralized, AI-powered networks.

## Features

- **Peer-to-Peer Networking**: Built on libp2p for robust P2P communication
- **Distributed Hash Table (DHT)**: Kademlia DHT for content and peer discovery
- **Publish-Subscribe Messaging**: GossipSub protocol for efficient message broadcasting
- **Configuration Management**: Flexible configuration system with sensible defaults
- **Connection Management**: Automatic connection management and NAT traversal
- **Extensible Architecture**: Modular design for easy extension and customization

## Quick Start

### Installation

```bash
go mod init your-project
go get github.com/olane-labs/olane-go
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/olane-labs/olane-go/pkg/config"
    "github.com/olane-labs/olane-go/pkg/node"
)

func main() {
    ctx := context.Background()

    // Create a node with default configuration
    n, err := node.NewNode(ctx, nil)
    if err != nil {
        log.Fatalf("Failed to create node: %v", err)
    }
    defer n.Stop()

    // Start the node
    if err := n.Start(); err != nil {
        log.Fatalf("Failed to start node: %v", err)
    }

    fmt.Printf("Node started with ID: %s\n", n.ID())
    fmt.Printf("Listening on: %v\n", n.Addrs())

    // Keep the node running
    select {}
}
```

### Custom Configuration

```go
package main

import (
    "context"
    "log"

    "github.com/olane-labs/olane-go/pkg/config"
    "github.com/olane-labs/olane-go/pkg/node"
)

func main() {
    ctx := context.Background()

    // Create custom configuration
    cfg := config.DefaultLibp2pConfig()
    cfg.Listeners = []string{"/ip4/0.0.0.0/tcp/4001"}
    cfg.BootstrapPeers = []string{
        "/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
    }

    // Create and start node
    n, err := node.NewNode(ctx, cfg)
    if err != nil {
        log.Fatalf("Failed to create node: %v", err)
    }
    defer n.Stop()

    if err := n.Start(); err != nil {
        log.Fatalf("Failed to start node: %v", err)
    }

    // Node is now running with custom configuration
    select {}
}
```

## Architecture

### Package Structure

```
olane-go/
├── cmd/
│   └── example/          # Example applications
├── pkg/
│   ├── config/          # Configuration management
│   ├── node/            # Node implementation
│   └── utils/           # Utility functions
└── go.mod
```

### Core Components

- **Config**: Manages libp2p configuration with sensible defaults
- **Node**: High-level node abstraction with lifecycle management
- **Utils**: Common utilities for key generation, address parsing, etc.

## API Reference

### Configuration

The `config` package provides configuration structures and factory functions:

```go
// Create default configuration
cfg := config.DefaultLibp2pConfig()

// Customize as needed
cfg.Listeners = []string{"/ip4/0.0.0.0/tcp/4001"}
cfg.EnableDHT = true
cfg.EnablePubsub = true
cfg.KBucketSize = 20
```

### Node Management

The `node` package provides high-level node management:

```go
// Create a new node
n, err := node.NewNode(ctx, cfg)

// Start the node
err = n.Start()

// Access node information
id := n.ID()
addrs := n.Addrs()
peers := n.Peers()

// Pub/Sub operations
sub, err := n.Subscribe("topic")
err = n.Publish(ctx, "topic", []byte("message"))

// DHT operations
err = n.PutValue(ctx, "key", []byte("value"))
value, err := n.GetValue(ctx, "key")

// Graceful shutdown
err = n.Stop()
```

### Utilities

The `utils` package provides helpful utility functions:

```go
// Generate key pairs
priv, pub, err := utils.GenerateKeyPair()
priv, pub, err := utils.GenerateEd25519KeyPair()

// Key serialization
b64, err := utils.PrivKeyToBase64(priv)
priv, err := utils.PrivKeyFromBase64(b64)

// Address validation
err := utils.ValidateMultiaddrs([]string{"/ip4/127.0.0.1/tcp/4001"})

// Configuration merging
merged := utils.MergeConfigs(baseConfig, overrideConfig)
```

## Examples

See the `cmd/example/` directory for complete examples:

- **Basic Node**: Simple node creation and startup
- **Custom Configuration**: Node with custom settings
- **Pub/Sub Messaging**: Publishing and subscribing to topics
- **DHT Storage**: Storing and retrieving data from the DHT

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Comparison with TypeScript Version

This Go implementation provides equivalent functionality to the TypeScript `@olane/o-config` package:

| TypeScript | Go |
|------------|-----|
| `defaultLibp2pConfig` | `config.DefaultLibp2pConfig()` |
| `createNode()` | `config.CreateNode()` / `node.NewNode()` |
| libp2p exports | Direct libp2p usage with Go types |
| Dynamic imports | Static imports with Go modules |

## Dependencies

- [libp2p](https://github.com/libp2p/go-libp2p): Core P2P networking
- [go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht): Kademlia DHT
- [go-libp2p-pubsub](https://github.com/libp2p/go-libp2p-pubsub): Publish-Subscribe messaging
- [go-multiaddr](https://github.com/multiformats/go-multiaddr): Multiaddress support

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the ISC License - see the LICENSE file for details.

## Roadmap

- [ ] Protocol buffer definitions for o-protocol
- [ ] Network bridge implementations
- [ ] Tool registry integration
- [ ] Advanced routing and discovery
- [ ] Performance optimizations
- [ ] Comprehensive documentation