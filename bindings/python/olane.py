"""
Python bindings for Olane Go implementation.

This module provides a Python interface to the Olane Go core and config functionality
using CGO bindings for high performance.
"""

import ctypes
import json
import os
from typing import Dict, List, Optional, Any
from pathlib import Path


class OlaneError(Exception):
    """Exception raised by Olane operations."""
    pass


class OAddress:
    """Python wrapper for Go OAddress functionality."""
    
    def __init__(self, address: str):
        self.address = address
        self._validate()
    
    def _validate(self):
        """Validate the address format."""
        if not self.address.startswith("o://"):
            raise OlaneError(f"Invalid address format: {self.address}")
    
    @property
    def root(self) -> str:
        """Get the root part of the address."""
        result = _lib.address_get_root(self.address.encode())
        return ctypes.string_at(result).decode()
    
    @property
    def paths(self) -> str:
        """Get the paths part of the address."""
        result = _lib.address_get_paths(self.address.encode())
        return ctypes.string_at(result).decode()
    
    @property
    def is_leader(self) -> bool:
        """Check if this is a leader address."""
        return bool(_lib.address_is_leader(self.address.encode()))
    
    @property
    def is_tool(self) -> bool:
        """Check if this is a tool address."""
        return bool(_lib.address_is_tool(self.address.encode()))
    
    def validate(self) -> bool:
        """Validate the address."""
        return bool(_lib.address_validate(self.address.encode()))
    
    def to_cid(self) -> str:
        """Convert address to CID."""
        result = _lib.address_get_cid(self.address.encode())
        cid_str = ctypes.string_at(result).decode()
        if cid_str.startswith("error:"):
            raise OlaneError(cid_str)
        return cid_str
    
    def __str__(self) -> str:
        return self.address
    
    def __repr__(self) -> str:
        return f"OAddress('{self.address}')"


class LibP2PConfig:
    """Python wrapper for Go LibP2P configuration."""
    
    def __init__(self, listeners: List[str] = None, enable_dht: bool = True, 
                 enable_pubsub: bool = True, enable_relay: bool = True,
                 k_bucket_size: int = 20):
        self.listeners = listeners or ["/ip4/0.0.0.0/tcp/0"]
        self.enable_dht = enable_dht
        self.enable_pubsub = enable_pubsub
        self.enable_relay = enable_relay
        self.k_bucket_size = k_bucket_size
    
    @classmethod
    def default(cls) -> 'LibP2PConfig':
        """Create default configuration from Go."""
        result = _lib.create_libp2p_config()
        config_json = ctypes.string_at(result).decode()
        config_data = json.loads(config_json)
        
        if "error" in config_data:
            raise OlaneError(config_data["error"])
        
        return cls(
            listeners=config_data.get("listeners", ["/ip4/0.0.0.0/tcp/0"]),
            enable_dht=config_data.get("enableDHT", True),
            enable_pubsub=config_data.get("enablePubsub", True),
            enable_relay=config_data.get("enableRelay", True),
            k_bucket_size=config_data.get("kBucketSize", 20)
        )
    
    def create_node(self) -> Dict[str, Any]:
        """Create a libp2p node with this configuration."""
        listeners_json = json.dumps(self.listeners)
        result = _lib.create_libp2p_node(listeners_json.encode())
        node_json = ctypes.string_at(result).decode()
        node_data = json.loads(node_json)
        
        if "error" in node_data:
            raise OlaneError(node_data["error"])
        
        return node_data


class CoreNode:
    """Python wrapper for Go CoreNode functionality."""
    
    def __init__(self, address: str, node_type: str = "node", 
                 name: str = "", description: str = ""):
        self.address = OAddress(address)
        self.node_type = node_type
        self.name = name
        self.description = description
        self._node_id = None
        self._is_started = False
    
    def _ensure_created(self):
        """Ensure the Go node is created."""
        if self._node_id is None:
            self._node_id = _lib.create_node(
                self.address.address.encode(),
                self.node_type.encode(),
                self.name.encode(),
                self.description.encode()
            )
    
    def start(self):
        """Start the node."""
        self._ensure_created()
        result = _lib.start_node(self._node_id)
        result_str = ctypes.string_at(result).decode()
        
        if result_str != "success":
            raise OlaneError(result_str)
        
        self._is_started = True
    
    def stop(self):
        """Stop the node."""
        if self._node_id is None:
            return
        
        result = _lib.stop_node(self._node_id)
        result_str = ctypes.string_at(result).decode()
        
        if result_str != "success":
            raise OlaneError(result_str)
        
        self._is_started = False
    
    def whoami(self) -> Dict[str, Any]:
        """Get node information."""
        self._ensure_created()
        
        result = _lib.node_whoami(self._node_id)
        whoami_json = ctypes.string_at(result).decode()
        whoami_data = json.loads(whoami_json)
        
        if "error" in whoami_data:
            raise OlaneError(whoami_data["error"])
        
        return whoami_data
    
    @property
    def is_started(self) -> bool:
        """Check if the node is started."""
        return self._is_started
    
    def __enter__(self):
        """Context manager entry."""
        self.start()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.stop()
        if self._node_id is not None:
            _lib.cleanup_node(self._node_id)
    
    def __del__(self):
        """Cleanup when object is destroyed."""
        if hasattr(self, '_node_id') and self._node_id is not None:
            try:
                if self._is_started:
                    self.stop()
                _lib.cleanup_node(self._node_id)
            except:
                pass  # Ignore errors during cleanup


# Load the shared library
def _load_library():
    """Load the Go shared library."""
    # Find library in the same directory as this Python file
    lib_dir = Path(__file__).parent
    lib_path = lib_dir / "lib_olane.so"
    
    if not lib_path.exists():
        # Try alternative names
        for name in ["libolane.so", "olane.so"]:
            alt_path = lib_dir / name
            if alt_path.exists():
                lib_path = alt_path
                break
        else:
            raise OlaneError(
                f"Could not find Olane shared library. "
                f"Expected at: {lib_path}\n"
                f"Run: cd {lib_dir} && ./build.sh"
            )
    
    try:
        lib = ctypes.CDLL(str(lib_path))
    except OSError as e:
        raise OlaneError(f"Failed to load shared library: {e}")
    
    # Define function signatures
    
    # Node functions
    lib.create_node.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
    lib.create_node.restype = ctypes.c_int
    
    lib.start_node.argtypes = [ctypes.c_int]
    lib.start_node.restype = ctypes.c_char_p
    
    lib.stop_node.argtypes = [ctypes.c_int]
    lib.stop_node.restype = ctypes.c_char_p
    
    lib.node_whoami.argtypes = [ctypes.c_int]
    lib.node_whoami.restype = ctypes.c_char_p
    
    lib.cleanup_node.argtypes = [ctypes.c_int]
    lib.cleanup_node.restype = None
    
    # Address functions
    lib.address_validate.argtypes = [ctypes.c_char_p]
    lib.address_validate.restype = ctypes.c_int
    
    lib.address_get_root.argtypes = [ctypes.c_char_p]
    lib.address_get_root.restype = ctypes.c_char_p
    
    lib.address_get_paths.argtypes = [ctypes.c_char_p]
    lib.address_get_paths.restype = ctypes.c_char_p
    
    lib.address_is_leader.argtypes = [ctypes.c_char_p]
    lib.address_is_leader.restype = ctypes.c_int
    
    lib.address_is_tool.argtypes = [ctypes.c_char_p]
    lib.address_is_tool.restype = ctypes.c_int
    
    lib.address_get_cid.argtypes = [ctypes.c_char_p]
    lib.address_get_cid.restype = ctypes.c_char_p
    
    # Config functions
    lib.create_libp2p_config.argtypes = []
    lib.create_libp2p_config.restype = ctypes.c_char_p
    
    lib.create_libp2p_node.argtypes = [ctypes.c_char_p]
    lib.create_libp2p_node.restype = ctypes.c_char_p
    
    # Utility functions
    lib.free_string.argtypes = [ctypes.c_char_p]
    lib.free_string.restype = None
    
    return lib


# Load library on import
try:
    _lib = _load_library()
except Exception as e:
    # If library loading fails, provide helpful error message
    _lib = None
    print(f"Warning: Could not load Olane Go library: {e}")
    print("To build the library, run: cd bindings/python && ./build.sh")


# Convenience functions
def create_address(address: str) -> OAddress:
    """Create an OAddress instance."""
    return OAddress(address)


def create_node(address: str, node_type: str = "node", 
                name: str = "", description: str = "") -> CoreNode:
    """Create a CoreNode instance."""
    return CoreNode(address, node_type, name, description)


def default_config() -> LibP2PConfig:
    """Get default libp2p configuration."""
    return LibP2PConfig.default()


# Version info
__version__ = "0.1.0"
__author__ = "Olane Labs"
__all__ = [
    "OAddress", "CoreNode", "LibP2PConfig", "OlaneError",
    "create_address", "create_node", "default_config"
]
