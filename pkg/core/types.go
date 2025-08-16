// Package core provides the core types and interfaces for Olane nodes.
package core

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"

	"github.com/olane-labs/olane-go/pkg/config"
)

// NodeState represents the current state of a node
type NodeState string

const (
	NodeStateStarting NodeState = "STARTING"
	NodeStateRunning  NodeState = "RUNNING"
	NodeStateStopping NodeState = "STOPPING"
	NodeStateStopped  NodeState = "STOPPED"
	NodeStateError    NodeState = "ERROR"
)

// NodeType represents the type of node
type NodeType string

const (
	NodeTypeLeader  NodeType = "leader"
	NodeTypeRoot    NodeType = "root"
	NodeTypeNode    NodeType = "node"
	NodeTypeTool    NodeType = "tool"
	NodeTypeAgent   NodeType = "agent"
	NodeTypeHuman   NodeType = "human"
	NodeTypeUnknown NodeType = "unknown"
)

// OMethod represents a method that can be called on a node
type OMethod struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Returns     map[string]interface{} `json:"returns"`
}

// ODependency represents a dependency of a node
type ODependency struct {
	Address     *OAddress `json:"address"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Optional    bool      `json:"optional"`
}

// CoreConfig holds the configuration for a core node
type CoreConfig struct {
	Address       *OAddress
	Leader        *OAddress
	Parent        *OAddress
	Type          NodeType
	Seed          string
	Name          string
	Network       *config.Libp2pConfig
	Metrics       bool
	Description   string
	Dependencies  []*ODependency
	Methods       map[string]*OMethod
	CWD           string
	NetworkName   string
	PromptAddress *OAddress
}

// DefaultCoreConfig returns a default core configuration
func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		Address:      NewOAddress("o://node"),
		Type:         NodeTypeUnknown,
		Network:      config.DefaultLibp2pConfig(),
		Metrics:      false,
		Dependencies: []*ODependency{},
		Methods:      make(map[string]*OMethod),
	}
}

// ORequest represents a request to a node
type ORequest struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// NewORequest creates a new ORequest
func NewORequest(id, method string, params map[string]interface{}) *ORequest {
	return &ORequest{
		ID:     id,
		Method: method,
		Params: params,
	}
}

// OResponse represents a response from a node
type OResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *OError     `json:"error,omitempty"`
}

// OError represents an error response
type OError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *OError) Error() string {
	return fmt.Sprintf("OError [%d]: %s", e.Code, e.Message)
}

// NewOResponse creates a new successful OResponse
func NewOResponse(id string, result interface{}) *OResponse {
	return &OResponse{
		ID:     id,
		Result: result,
	}
}

// NewOErrorResponse creates a new error OResponse
func NewOErrorResponse(id string, code int, message string, data interface{}) *OResponse {
	return &OResponse{
		ID: id,
		Error: &OError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// UseOptions provides options for the Use method
type UseOptions struct {
	NoIndex bool
	Timeout int // timeout in seconds
}

// DefaultUseOptions returns default use options
func DefaultUseOptions() *UseOptions {
	return &UseOptions{
		NoIndex: false,
		Timeout: 30,
	}
}

// ConnectionSendParams represents parameters for sending data over a connection
type ConnectionSendParams struct {
	Address string                 `json:"address"`
	Payload map[string]interface{} `json:"payload"`
}

// WhoAmIResponse represents the response from the whoami method
type WhoAmIResponse struct {
	Address      string              `json:"address"`
	Type         NodeType            `json:"type"`
	Description  string              `json:"description"`
	Methods      map[string]*OMethod `json:"methods"`
	SuccessCount int64               `json:"successCount"`
	ErrorCount   int64               `json:"errorCount"`
	PeerID       string              `json:"peerId"`
	Transports   []string            `json:"transports"`
}

// Logger interface for structured logging
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Connection interface represents a connection to another node
type Connection interface {
	Send(ctx context.Context, params *ConnectionSendParams) (*OResponse, error)
	Close() error
	RemotePeer() peer.ID
	RemoteAddr() multiaddr.Multiaddr
}

// ConnectionManager interface manages connections to other nodes
type ConnectionManager interface {
	Connect(ctx context.Context, params *ConnectionParams) (Connection, error)
	Disconnect(peerID peer.ID) error
	GetConnection(peerID peer.ID) (Connection, bool)
	ListConnections() []Connection
}

// ConnectionParams represents parameters for establishing a connection
type ConnectionParams struct {
	Address       *OAddress
	NextHopAddress *OAddress
	CallerAddress *OAddress
}

// AddressResolver interface for resolving addresses
type AddressResolver interface {
	Resolve(ctx context.Context, address *OAddress) (*OAddress, error)
	SupportsTransport(address *OAddress) bool
}

// AddressResolution manages multiple address resolvers
type AddressResolution struct {
	resolvers []AddressResolver
}

// NewAddressResolution creates a new address resolution manager
func NewAddressResolution() *AddressResolution {
	return &AddressResolution{
		resolvers: make([]AddressResolver, 0),
	}
}

// AddResolver adds a new address resolver
func (ar *AddressResolution) AddResolver(resolver AddressResolver) {
	ar.resolvers = append(ar.resolvers, resolver)
}

// Resolve resolves an address using the available resolvers
func (ar *AddressResolution) Resolve(ctx context.Context, address *OAddress) (*OAddress, error) {
	for _, resolver := range ar.resolvers {
		if resolved, err := resolver.Resolve(ctx, address); err == nil {
			return resolved, nil
		}
	}
	// If no resolver can handle it, return the original address
	return address, nil
}

// SupportsTransport checks if any resolver supports the given transport
func (ar *AddressResolution) SupportsTransport(address *OAddress) bool {
	for _, resolver := range ar.resolvers {
		if resolver.SupportsTransport(address) {
			return true
		}
	}
	return false
}

// CIDProvider interface for generating CIDs
type CIDProvider interface {
	ToCID(data interface{}) (multihash.Multihash, error)
}

// TranslateAddressResult represents the result of address translation
type TranslateAddressResult struct {
	NextHopAddress *OAddress
	TargetAddress  *OAddress
}

// NodeInterface defines the interface that all node types must implement
type NodeInterface interface {
	// Core lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Initialize(ctx context.Context) error

	// Identity and information
	ID() peer.ID
	Address() *OAddress
	Type() NodeType
	WhoAmI(ctx context.Context) (*WhoAmIResponse, error)

	// Network operations
	Use(ctx context.Context, address *OAddress, method string, params map[string]interface{}, opts *UseOptions) (*OResponse, error)
	Connect(ctx context.Context, nextHopAddress, targetAddress *OAddress) (Connection, error)

	// State management
	State() NodeState
	Errors() []error

	// Network management
	Register(ctx context.Context) error
	Unregister(ctx context.Context) error
	AdvertiseToNetwork(ctx context.Context) error

	// Transport and addressing
	Transports() []string
	GetTransports(address *OAddress) []multiaddr.Multiaddr
	TranslateAddress(ctx context.Context, address *OAddress) (*TranslateAddressResult, error)

	// Libp2p integration
	Host() host.Host
}
