#!/usr/bin/env python3
"""
Example demonstrating Python bindings for Olane Go config package.

This example shows how to use the Go pkg/config functionality from Python
for libp2p configuration and node creation.
"""

import json
import time
import signal
import sys
from olane_config import (
    Libp2pConfig, Libp2pNode, OlaneConfigError,
    default_config, create_node, validate_multiaddr, validate_multiaddrs
)


def config_examples():
    """Demonstrate Libp2pConfig functionality."""
    print("⚙️  === Configuration Examples ===")
    
    # Get default configuration
    print("\n1. Default Configuration:")
    try:
        config = default_config()
        print(f"  Listeners: {config.listeners}")
        print(f"  Bootstrap peers: {config.bootstrap_peers}")
        print(f"  DHT enabled: {config.enable_dht}")
        print(f"  PubSub enabled: {config.enable_pubsub}")
        print(f"  Relay enabled: {config.enable_relay}")
        print(f"  K-bucket size: {config.k_bucket_size}")
        
    except OlaneConfigError as e:
        print(f"❌ Config error: {e}")
    
    # Create custom configuration
    print("\n2. Custom Configuration:")
    try:
        custom_config = Libp2pConfig(
            listeners=["/ip4/0.0.0.0/tcp/4001", "/ip4/127.0.0.1/tcp/4002"],
            bootstrap_peers=[
                "/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"
            ],
            enable_dht=True,
            enable_pubsub=True,
            enable_relay=False,
            k_bucket_size=30
        )
        
        print(f"  Custom listeners: {custom_config.listeners}")
        print(f"  Bootstrap peers: {len(custom_config.bootstrap_peers)} configured")
        print(f"  Relay disabled: {not custom_config.enable_relay}")
        print(f"  Larger k-bucket: {custom_config.k_bucket_size}")
        
        # Test JSON serialization
        config_json = custom_config.to_json()
        print(f"  JSON size: {len(config_json)} bytes")
        
        # Test JSON deserialization
        restored_config = Libp2pConfig.from_json(config_json)
        print(f"  Restored config: {restored_config}")
        
    except OlaneConfigError as e:
        print(f"❌ Custom config error: {e}")


def address_validation_examples():
    """Demonstrate multiaddr validation."""
    print("\n🔍 === Address Validation Examples ===")
    
    test_addresses = [
        "/ip4/127.0.0.1/tcp/4001",
        "/ip4/0.0.0.0/tcp/0",
        "/ip6/::1/tcp/4001",
        "/ip4/192.168.1.100/tcp/8080/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
        "invalid-address",
        "/invalid/protocol/test",
    ]
    
    print("\n1. Individual Address Validation:")
    for addr in test_addresses:
        is_valid = validate_multiaddr(addr)
        status = "✅ Valid" if is_valid else "❌ Invalid"
        print(f"  {addr:<70} {status}")
    
    print("\n2. Batch Address Validation:")
    valid_addrs = [addr for addr in test_addresses if validate_multiaddr(addr)]
    invalid_addrs = [addr for addr in test_addresses if not validate_multiaddr(addr)]
    
    print(f"  Valid addresses ({len(valid_addrs)}):")
    for addr in valid_addrs:
        print(f"    {addr}")
    
    print(f"  Invalid addresses ({len(invalid_addrs)}):")
    for addr in invalid_addrs:
        print(f"    {addr}")
    
    # Test batch validation
    all_valid = validate_multiaddrs(valid_addrs)
    print(f"\n  Batch validation of valid addresses: {all_valid}")
    
    mixed_valid = validate_multiaddrs(test_addresses)
    print(f"  Batch validation of mixed addresses: {mixed_valid}")


def node_creation_examples():
    """Demonstrate libp2p node creation and management."""
    print("\n🌐 === Node Creation Examples ===")
    
    # Create node with default configuration
    print("\n1. Default Node Creation:")
    try:
        with create_node() as node:
            print(f"  ✅ Node created successfully")
            print(f"  Peer ID: {node.peer_id}")
            print(f"  Has DHT: {node.has_dht}")
            print(f"  Has PubSub: {node.has_pubsub}")
            print(f"  Protocols: {len(node.protocols)} available")
            print(f"  Listen addresses: {len(node.addrs)}")
            
            for i, addr in enumerate(node.addrs[:3]):  # Show first 3 addresses
                print(f"    {i+1}. {addr}")
            
            if len(node.addrs) > 3:
                print(f"    ... and {len(node.addrs) - 3} more")
        
        print("  ✅ Node automatically closed")
        
    except OlaneConfigError as e:
        print(f"  ❌ Node creation error: {e}")
    
    # Create node with custom configuration
    print("\n2. Custom Node Creation:")
    try:
        config = Libp2pConfig(
            listeners=["/ip4/127.0.0.1/tcp/0"],  # Random port on localhost
            enable_dht=True,
            enable_pubsub=False,  # Disable pubsub for this example
            k_bucket_size=10
        )
        
        node = config.create_node()
        print(f"  ✅ Custom node created")
        print(f"  Peer ID: {node.peer_id}")
        print(f"  Has DHT: {node.has_dht}")
        print(f"  Has PubSub: {node.has_pubsub}")
        
        # Get detailed node information
        info = node.get_info()
        print(f"  Connected peers: {info['peerCount']}")
        print(f"  Protocols: {len(info['protocols'])}")
        
        # Close manually
        node.close()
        print(f"  ✅ Node closed manually")
        
    except OlaneConfigError as e:
        print(f"  ❌ Custom node error: {e}")


def multi_node_example():
    """Demonstrate multiple nodes and connections."""
    print("\n🔗 === Multi-Node Example ===")
    
    nodes = []
    try:
        # Create multiple nodes
        print("Creating 3 nodes...")
        for i in range(3):
            config = Libp2pConfig(
                listeners=[f"/ip4/127.0.0.1/tcp/{4000 + i}"],
                enable_dht=True,
                enable_pubsub=True
            )
            
            node = config.create_node()
            nodes.append(node)
            print(f"  Node {i+1}: {node.peer_id[:16]}... on port {4000 + i}")
        
        # Wait a moment for nodes to initialize
        time.sleep(2)
        
        # Show node information
        print("\nNode information:")
        for i, node in enumerate(nodes):
            info = node.get_info()
            peer_count = node.get_peer_count()
            print(f"  Node {i+1}: {peer_count} peers, {len(info['protocols'])} protocols")
        
        # Try to connect nodes (would need bootstrap configuration in real scenario)
        print("\nNote: For peer discovery, configure bootstrap peers or use mDNS")
        
    except OlaneConfigError as e:
        print(f"❌ Multi-node error: {e}")
    
    finally:
        # Clean up all nodes
        print("\nCleaning up nodes...")
        for i, node in enumerate(nodes):
            try:
                node.close()
                print(f"  ✅ Node {i+1} closed")
            except:
                print(f"  ⚠️  Node {i+1} cleanup error")


def monitoring_example():
    """Demonstrate node monitoring capabilities."""
    print("\n📊 === Node Monitoring Example ===")
    
    try:
        config = Libp2pConfig(
            listeners=["/ip4/0.0.0.0/tcp/0"],
            enable_dht=True,
            enable_pubsub=True
        )
        
        with config.create_node() as node:
            print(f"Monitoring node: {node.peer_id[:16]}...")
            
            # Monitor for 10 seconds
            for i in range(10):
                time.sleep(1)
                
                peer_count = node.get_peer_count()
                connected_peers = node.get_connected_peers()
                
                print(f"  T+{i+1}s: {peer_count} peers connected")
                
                if connected_peers:
                    print(f"    Connected to: {[p[:16] + '...' for p in connected_peers[:3]]}")
                
                # Get detailed info every 3 seconds
                if (i + 1) % 3 == 0:
                    info = node.get_info()
                    print(f"    Protocols: {len(info['protocols'])}, Addresses: {len(info['addrs'])}")
            
            print("  📊 Monitoring completed")
        
    except OlaneConfigError as e:
        print(f"❌ Monitoring error: {e}")
    except KeyboardInterrupt:
        print("\n⚠️  Monitoring interrupted")


def main():
    """Run all examples."""
    print("🌊 Olane Config Python-Go Integration Demo")
    print("=" * 60)
    
    # Setup signal handler for graceful shutdown
    def signal_handler(sig, frame):
        print("\n\n🛑 Received interrupt signal, shutting down...")
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    
    try:
        # Run examples
        config_examples()
        address_validation_examples()
        node_creation_examples()
        multi_node_example()
        monitoring_example()
        
    except Exception as e:
        print(f"\n❌ Demo failed: {e}")
        import traceback
        traceback.print_exc()
        return 1
    
    print("\n🎉 All config examples completed successfully!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
