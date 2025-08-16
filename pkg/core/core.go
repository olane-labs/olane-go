// Package core provides the core types and functionality for Olane networks.
//
// This package contains the fundamental building blocks for creating Olane nodes:
//
//   - OAddress: Represents o-protocol addresses (o://)
//   - CoreNode: Base implementation for all node types
//   - NodeInterface: Interface that all nodes must implement
//   - Connection management and address resolution
//   - Logging and error handling utilities
//
// The core package is designed to be extended by concrete node implementations
// such as HostNode, VirtualNode, etc.
//
// Example usage:
//
//	// Create a core configuration
//	cfg := core.DefaultCoreConfig()
//	cfg.Address = core.NewOAddress("o://my-node")
//	cfg.Type = core.NodeTypeNode
//
//	// Create a node
//	node := core.NewCoreNode(cfg)
//
//	// Start the node
//	ctx := context.Background()
//	if err := node.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//
//	// Use the node to communicate with other nodes
//	response, err := node.Use(ctx, 
//		core.NewOAddress("o://other-node"), 
//		"method", 
//		map[string]interface{}{"param": "value"},
//		nil)
package core

const (
	// Version is the current version of the core package
	Version = "0.1.0"
	
	// ProtocolVersion is the o-protocol version supported
	ProtocolVersion = "1.0.0"
	
	// DefaultTimeout is the default timeout for operations
	DefaultTimeout = 30 // seconds
)

// Core error codes
const (
	ErrorCodeGeneral           = 1000
	ErrorCodeInvalidAddress    = 1001
	ErrorCodeConnectionFailed  = 1002
	ErrorCodeNodeNotRunning    = 1003
	ErrorCodeMethodNotFound    = 1004
	ErrorCodeTimeout           = 1005
	ErrorCodeInvalidResponse   = 1006
	ErrorCodeRegistrationFailed = 1007
)

// NewOError creates a new OError with the given code and message
func NewOError(code int, message string, data interface{}) *OError {
	return &OError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Common error constructors
var (
	ErrInvalidAddress = func(addr string) *OError {
		return NewOError(ErrorCodeInvalidAddress, "invalid address: "+addr, nil)
	}
	
	ErrConnectionFailed = func(target string, cause error) *OError {
		return NewOError(ErrorCodeConnectionFailed, "connection failed to "+target, cause.Error())
	}
	
	ErrNodeNotRunning = func() *OError {
		return NewOError(ErrorCodeNodeNotRunning, "node is not running", nil)
	}
	
	ErrMethodNotFound = func(method string) *OError {
		return NewOError(ErrorCodeMethodNotFound, "method not found: "+method, nil)
	}
	
	ErrTimeout = func(operation string) *OError {
		return NewOError(ErrorCodeTimeout, "operation timed out: "+operation, nil)
	}
)

// ProtocolInfo contains information about the o-protocol
type ProtocolInfo struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

// GetProtocolInfo returns information about the o-protocol
func GetProtocolInfo() *ProtocolInfo {
	return &ProtocolInfo{
		Version: ProtocolVersion,
		Name:    "o-protocol",
	}
}

// IsValidNodeType checks if a node type is valid
func IsValidNodeType(nodeType NodeType) bool {
	switch nodeType {
	case NodeTypeLeader, NodeTypeRoot, NodeTypeNode, NodeTypeTool, NodeTypeAgent, NodeTypeHuman, NodeTypeUnknown:
		return true
	default:
		return false
	}
}

// IsValidNodeState checks if a node state is valid
func IsValidNodeState(state NodeState) bool {
	switch state {
	case NodeStateStarting, NodeStateRunning, NodeStateStopping, NodeStateStopped, NodeStateError:
		return true
	default:
		return false
	}
}
