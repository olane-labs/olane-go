#!/usr/bin/env python3
"""
Example demonstrating Python bindings for Olane Go implementation.

This example shows how to use the Go core and config functionality from Python
using CGO bindings.
"""

import json
import time
import signal
import sys
from olane import (
    OAddress, CoreNode, LibP2PConfig, OlaneError,
    create_address, create_node, default_config
)


def address_examples():
    """Demonstrate OAddress functionality."""
    print("🔗 === Address Examples ===")
    
    addresses = [
        "o://leader",
        "o://tools/calculator", 
        "o://services/weather/current",
        "o://ai/gpt/chat"
    ]
    
    for addr_str in addresses:
        addr = create_address(addr_str)
        print(f"\nAddress: {addr}")
        print(f"  Root: {addr.root}")
        print(f"  Paths: {addr.paths}")
        print(f"  Is Leader: {addr.is_leader}")
        print(f"  Is Tool: {addr.is_tool}")
        print(f"  Valid: {addr.validate()}")
        
        try:
            cid = addr.to_cid()
            print(f"  CID: {cid}")
        except OlaneError as e:
            print(f"  CID Error: {e}")


def config_examples():
    """Demonstrate LibP2P configuration."""
    print("\n⚙️  === Configuration Examples ===")
    
    # Get default configuration
    try:
        config = default_config()
        print(f"Default listeners: {config.listeners}")
        print(f"DHT enabled: {config.enable_dht}")
        print(f"PubSub enabled: {config.enable_pubsub}")
        print(f"Relay enabled: {config.enable_relay}")
        print(f"K-bucket size: {config.k_bucket_size}")
        
        # Create a libp2p node (this would need actual network setup)
        # node_info = config.create_node()
        # print(f"Node info: {json.dumps(node_info, indent=2)}")
        
    except OlaneError as e:
        print(f"Config error: {e}")


def node_examples():
    """Demonstrate CoreNode functionality."""
    print("\n🌐 === Node Examples ===")
    
    # Create nodes
    nodes = [
        create_node("o://example-node-1", "node", "node1", "First example node"),
        create_node("o://example-node-2", "tool", "node2", "Second example node"),
        create_node("o://leader-node", "leader", "leader", "Leader node"),
    ]
    
    for node in nodes:
        print(f"\nNode: {node.address}")
        print(f"  Type: {node.node_type}")
        print(f"  Name: {node.name}")
        print(f"  Description: {node.description}")
        
        try:
            # Start the node
            print("  Starting node...")
            node.start()
            print("  ✅ Node started successfully")
            
            # Get node information
            whoami = node.whoami()
            print(f"  Address: {whoami['address']}")
            print(f"  Type: {whoami['type']}")
            print(f"  Success Count: {whoami['successCount']}")
            print(f"  Error Count: {whoami['errorCount']}")
            print(f"  Methods: {len(whoami['methods'])} available")
            
            # Stop the node
            print("  Stopping node...")
            node.stop()
            print("  ✅ Node stopped successfully")
            
        except OlaneError as e:
            print(f"  ❌ Node error: {e}")


def context_manager_example():
    """Demonstrate using nodes with context managers."""
    print("\n🔄 === Context Manager Example ===")
    
    try:
        with create_node("o://context-node", "agent", "context", "Context managed node") as node:
            print(f"Node {node.address} is running in context")
            whoami = node.whoami()
            print(f"Node type: {whoami['type']}")
            print(f"Node description: {whoami['description']}")
            
        print("✅ Node automatically stopped when exiting context")
        
    except OlaneError as e:
        print(f"❌ Context error: {e}")


def interactive_demo():
    """Interactive demonstration."""
    print("\n🎮 === Interactive Demo ===")
    
    # Setup signal handler for graceful shutdown
    def signal_handler(sig, frame):
        print("\n\n🛑 Received interrupt signal, shutting down...")
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    
    print("Creating and starting a demo node...")
    
    try:
        node = create_node("o://interactive-demo", "node", "demo", "Interactive demo node")
        node.start()
        
        print(f"✅ Demo node {node.address} is running!")
        print("📊 Node statistics:")
        
        # Monitor the node
        for i in range(5):
            time.sleep(1)
            whoami = node.whoami()
            print(f"  Iteration {i+1}: Success={whoami['successCount']}, Errors={whoami['errorCount']}")
        
        print("\n🛑 Stopping demo node...")
        node.stop()
        print("✅ Demo completed successfully!")
        
    except OlaneError as e:
        print(f"❌ Demo error: {e}")
    except KeyboardInterrupt:
        print("\n🛑 Demo interrupted by user")


def main():
    """Run all examples."""
    print("🌊 Olane Python-Go Integration Demo")
    print("=" * 50)
    
    try:
        # Run examples
        address_examples()
        config_examples()
        node_examples()
        context_manager_example()
        interactive_demo()
        
    except Exception as e:
        print(f"\n❌ Demo failed: {e}")
        import traceback
        traceback.print_exc()
        return 1
    
    print("\n🎉 All examples completed successfully!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
