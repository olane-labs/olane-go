"""
Python bindings for Olane Go config package.

This module provides Python access to the olane-go pkg/config functionality
for libp2p configuration and node creation using CGO bindings.
"""

import ctypes
import json
import os
from typing import Dict, List, Optional, Any
from pathlib import Path


class OlaneConfigError(Exception):
    """Exception raised by Olane config operations."""
    pass


class Libp2pConfig:
    """Python wrapper for Go Libp2pConfig functionality."""
    
    def __init__(self, listeners: List[str] = None, bootstrap_peers: List[str] = None,
                 enable_relay: bool = True, enable_dht: bool = True, 
                 enable_pubsub: bool = True, k_bucket_size: int = 20):
        """Initialize libp2p configuration.
        
        Args:
            listeners: List of multiaddr strings to listen on
            bootstrap_peers: List of bootstrap peer multiaddrs
            enable_relay: Enable circuit relay functionality
            enable_dht: Enable Kademlia DHT
            enable_pubsub: Enable GossipSub
            k_bucket_size: DHT k-bucket size
        """
        self.listeners = listeners or ["/ip4/0.0.0.0/tcp/0"]
        self.bootstrap_peers = bootstrap_peers or []
        self.enable_relay = enable_relay
        self.enable_dht = enable_dht
        self.enable_pubsub = enable_pubsub
        self.k_bucket_size = k_bucket_size
    
    @classmethod
    def default(cls) -> 'Libp2pConfig':
        """Create default configuration from Go."""
        result = _lib.get_default_config()
        if not result:
            raise OlaneConfigError("get_default_config returned NULL")
        config_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        config_data = json.loads(config_json)
        if "error" in config_data:
            raise OlaneConfigError(config_data["error"])
        
        return cls(
            listeners=config_data.get("listeners", ["/ip4/0.0.0.0/tcp/0"]),
            bootstrap_peers=config_data.get("bootstrapPeers", []),
            enable_relay=config_data.get("enableRelay", True),
            enable_dht=config_data.get("enableDHT", True),
            enable_pubsub=config_data.get("enablePubsub", True),
            k_bucket_size=config_data.get("kBucketSize", 20)
        )
    
    def to_json(self) -> str:
        """Convert configuration to JSON string."""
        return json.dumps({
            "listeners": self.listeners,
            "bootstrapPeers": self.bootstrap_peers,
            "enableRelay": self.enable_relay,
            "enableDHT": self.enable_dht,
            "enablePubsub": self.enable_pubsub,
            "kBucketSize": self.k_bucket_size
        })
    
    @classmethod
    def from_json(cls, json_str: str) -> 'Libp2pConfig':
        """Create configuration from JSON string."""
        data = json.loads(json_str)
        return cls(
            listeners=data.get("listeners"),
            bootstrap_peers=data.get("bootstrapPeers"),
            enable_relay=data.get("enableRelay", True),
            enable_dht=data.get("enableDHT", True),
            enable_pubsub=data.get("enablePubsub", True),
            k_bucket_size=data.get("kBucketSize", 20)
        )
    
    def create_node(self) -> 'Libp2pNode':
        """Create a libp2p node with this configuration."""
        return Libp2pNode.create(self)
    
    def __repr__(self) -> str:
        return f"Libp2pConfig(listeners={len(self.listeners)}, dht={self.enable_dht}, pubsub={self.enable_pubsub})"


class Libp2pNode:
    """Python wrapper for a Go libp2p node."""
    
    def __init__(self, node_id: int, peer_id: str, addrs: List[str], 
                 has_dht: bool, has_pubsub: bool, protocols: List[str]):
        """Initialize node wrapper (use create() class method instead)."""
        self.node_id = node_id
        self.peer_id = peer_id
        self.addrs = addrs
        self.has_dht = has_dht
        self.has_pubsub = has_pubsub
        self.protocols = protocols
        self._closed = False
    
    @classmethod
    def create(cls, config: Libp2pConfig) -> 'Libp2pNode':
        """Create a new libp2p node with the given configuration."""
        config_json = config.to_json()
        result = _lib.create_node(config_json.encode())
        node_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        node_data = json.loads(node_json)
        if "error" in node_data:
            raise OlaneConfigError(node_data["error"])
        
        return cls(
            node_id=node_data["id"],
            peer_id=node_data["peerId"],
            addrs=node_data["addrs"],
            has_dht=node_data["hasDHT"],
            has_pubsub=node_data["hasPubsub"],
            protocols=node_data["protocols"]
        )
    
    def connect_to_bootstrap_peers(self, bootstrap_peers: List[str]) -> None:
        """Connect to bootstrap peers."""
        if self._closed:
            raise OlaneConfigError("Node is closed")
        
        peers_json = json.dumps(bootstrap_peers)
        result = _lib.connect_to_bootstrap_peers(self.node_id, peers_json.encode())
        response_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        response = json.loads(response_json)
        if "error" in response:
            raise OlaneConfigError(response["error"])
    
    def get_info(self) -> Dict[str, Any]:
        """Get current node information."""
        if self._closed:
            raise OlaneConfigError("Node is closed")
        
        result = _lib.get_node_info(self.node_id)
        info_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        info_data = json.loads(info_json)
        if "error" in info_data:
            raise OlaneConfigError(info_data["error"])
        
        return info_data
    
    def get_peer_count(self) -> int:
        """Get the number of connected peers."""
        if self._closed:
            return 0
        
        count = _lib.get_peer_count(self.node_id)
        return max(0, int(count))  # Return 0 if negative (error)
    
    def get_connected_peers(self) -> List[str]:
        """Get list of connected peer IDs."""
        if self._closed:
            return []
        
        result = _lib.get_connected_peers(self.node_id)
        peers_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        peers_data = json.loads(peers_json)
        if "error" in peers_data:
            raise OlaneConfigError(peers_data["error"])
        
        return peers_data["peers"]
    
    def close(self) -> None:
        """Close the node and free resources."""
        if self._closed:
            return
        
        result = _lib.close_node(self.node_id)
        response_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        response = json.loads(response_json)
        if "error" in response:
            raise OlaneConfigError(response["error"])
        
        self._closed = True
    
    def is_closed(self) -> bool:
        """Check if the node is closed."""
        return self._closed
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()
    
    def __del__(self):
        """Cleanup when object is destroyed."""
        if not self._closed:
            try:
                self.close()
            except:
                pass  # Ignore errors during cleanup
    
    def __repr__(self) -> str:
        status = "closed" if self._closed else "open"
        return f"Libp2pNode(peer_id='{self.peer_id[:16]}...', status={status})"


# Utility functions
def validate_multiaddr(addr: str) -> bool:
    """Validate a single multiaddr string."""
    result = _lib.validate_multiaddr(addr.encode())
    return bool(result)


def validate_multiaddrs(addrs: List[str]) -> bool:
    """Validate a list of multiaddr strings."""
    try:
        addrs_json = json.dumps(addrs)
        result = _lib.validate_multiaddrs(addrs_json.encode())
        response_json = ctypes.string_at(result).decode()
        # Note: We don't free the string to avoid memory management issues
        
        response = json.loads(response_json)
        if "error" in response:
            return False
        
        return response.get("valid", False)
    except:
        return False


# Load the shared library
def _load_library():
    """Load the Go config shared library."""
    # Find library in the same directory as this Python file
    lib_dir = Path(__file__).parent
    lib_path = lib_dir / "lib_olane_config.so"
    
    if not lib_path.exists():
        # Try alternative names
        for name in ["libolane_config.so", "olane_config.so"]:
            alt_path = lib_dir / name
            if alt_path.exists():
                lib_path = alt_path
                break
        else:
            raise OlaneConfigError(
                f"Could not find Olane config shared library. "
                f"Expected at: {lib_path}\n"
                f"Run: cd {lib_dir} && ./build.sh"
            )
    
    try:
        lib = ctypes.CDLL(str(lib_path))
    except OSError as e:
        raise OlaneConfigError(f"Failed to load shared library: {e}")
    
    # Define function signatures
    
    # Config functions
    lib.get_default_config.argtypes = []
    lib.get_default_config.restype = ctypes.c_char_p
    
    lib.create_config.argtypes = [
        ctypes.c_char_p,  # listeners JSON
        ctypes.c_char_p,  # bootstrap peers JSON
        ctypes.c_int,     # enable relay
        ctypes.c_int,     # enable DHT
        ctypes.c_int,     # enable pubsub
        ctypes.c_int      # k bucket size
    ]
    lib.create_config.restype = ctypes.c_char_p
    
    # Node functions
    lib.create_node.argtypes = [ctypes.c_char_p]
    lib.create_node.restype = ctypes.c_char_p
    
    lib.connect_to_bootstrap_peers.argtypes = [ctypes.c_int, ctypes.c_char_p]
    lib.connect_to_bootstrap_peers.restype = ctypes.c_char_p
    
    lib.get_node_info.argtypes = [ctypes.c_int]
    lib.get_node_info.restype = ctypes.c_char_p
    
    lib.close_node.argtypes = [ctypes.c_int]
    lib.close_node.restype = ctypes.c_char_p
    
    lib.get_peer_count.argtypes = [ctypes.c_int]
    lib.get_peer_count.restype = ctypes.c_int
    
    lib.get_connected_peers.argtypes = [ctypes.c_int]
    lib.get_connected_peers.restype = ctypes.c_char_p
    
    # Validation functions
    lib.validate_multiaddr.argtypes = [ctypes.c_char_p]
    lib.validate_multiaddr.restype = ctypes.c_int
    
    lib.validate_multiaddrs.argtypes = [ctypes.c_char_p]
    lib.validate_multiaddrs.restype = ctypes.c_char_p
    
    # Note: No free_string function to avoid memory management issues
    
    return lib


# Load library on import
try:
    _lib = _load_library()
except Exception as e:
    # If library loading fails, provide helpful error message
    _lib = None
    print(f"Warning: Could not load Olane config library: {e}")
    print("To build the library, run: cd bindings/config && ./build.sh")


# Convenience functions
def default_config() -> Libp2pConfig:
    """Get default libp2p configuration."""
    return Libp2pConfig.default()


def create_node(config: Libp2pConfig = None) -> Libp2pNode:
    """Create a libp2p node with optional configuration."""
    if config is None:
        config = default_config()
    return config.create_node()


# Version info
__version__ = "0.1.0"
__author__ = "Olane Labs"
__all__ = [
    "Libp2pConfig", "Libp2pNode", "OlaneConfigError",
    "default_config", "create_node", 
    "validate_multiaddr", "validate_multiaddrs"
]
