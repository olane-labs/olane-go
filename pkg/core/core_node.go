package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/olane-labs/olane-go/pkg/config"
)

// CoreNode is the base implementation of a node in the Olane network
type CoreNode struct {
	// Core properties
	p2pNode           host.Host
	logger            Logger
	networkConfig     *config.Libp2pConfig
	address           *OAddress
	staticAddress     *OAddress
	peerId            peer.ID
	state             NodeState
	errors            []error
	connectionManager ConnectionManager
	leaders           []multiaddr.Multiaddr
	addressResolution *AddressResolution
	description       string
	dependencies      []*ODependency
	methods           map[string]*OMethod

	// Statistics
	successCount int64
	errorCount   int64

	// Configuration
	config *CoreConfig

	// Synchronization
	mu sync.RWMutex
}

// NewCoreNode creates a new CoreNode with the given configuration
func NewCoreNode(cfg *CoreConfig) *CoreNode {
	if cfg == nil {
		cfg = DefaultCoreConfig()
	}

	// Create logger name with optional node name
	loggerName := "CoreNode"
	if cfg.Name != "" {
		loggerName = fmt.Sprintf("CoreNode:%s", cfg.Name)
	}
	loggerName = fmt.Sprintf("%s:%s", loggerName, cfg.Address.String())

	node := &CoreNode{
		logger:            NewLogger(loggerName),
		address:           cfg.Address,
		staticAddress:     cfg.Address.Clone(),
		networkConfig:     cfg.Network,
		state:             NodeStateStopped,
		errors:            make([]error, 0),
		addressResolution: NewAddressResolution(),
		description:       cfg.Description,
		dependencies:      cfg.Dependencies,
		methods:           cfg.Methods,
		config:            cfg,
		successCount:      0,
		errorCount:        0,
	}

	if node.networkConfig == nil {
		node.networkConfig = config.DefaultLibp2pConfig()
	}

	if node.methods == nil {
		node.methods = make(map[string]*OMethod)
	}

	if node.dependencies == nil {
		node.dependencies = make([]*ODependency, 0)
	}

	return node
}

// Type returns the node type
func (n *CoreNode) Type() NodeType {
	if n.config.Type == "" {
		return NodeTypeUnknown
	}
	return n.config.Type
}

// State returns the current node state
func (n *CoreNode) State() NodeState {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state
}

// setState sets the node state (thread-safe)
func (n *CoreNode) setState(state NodeState) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.state = state
}

// ID returns the peer ID of the node
func (n *CoreNode) ID() peer.ID {
	return n.peerId
}

// Address returns the node's address
func (n *CoreNode) Address() *OAddress {
	return n.address
}

// Host returns the libp2p host
func (n *CoreNode) Host() host.Host {
	return n.p2pNode
}

// Errors returns the list of errors that occurred
func (n *CoreNode) Errors() []error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]error, len(n.errors))
	copy(result, n.errors)
	return result
}

// addError adds an error to the error list (thread-safe)
func (n *CoreNode) addError(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.errors = append(n.errors, err)
}

// Transports returns the multiaddresses this node is listening on
func (n *CoreNode) Transports() []string {
	if n.p2pNode == nil {
		return []string{}
	}
	
	addrs := n.p2pNode.Addrs()
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

// WhoAmI returns information about this node
func (n *CoreNode) WhoAmI(ctx context.Context) (*WhoAmIResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return &WhoAmIResponse{
		Address:      n.address.String(),
		Type:         n.Type(),
		Description:  n.description,
		Methods:      n.methods,
		SuccessCount: n.successCount,
		ErrorCount:   n.errorCount,
		PeerID:       n.peerId.String(),
		Transports:   n.Transports(),
	}, nil
}

// Parent returns the parent address if configured
func (n *CoreNode) Parent() *OAddress {
	return n.config.Parent
}

// ParentPeerID extracts the peer ID from the parent's transport
func (n *CoreNode) ParentPeerID() (peer.ID, error) {
	parent := n.Parent()
	if parent == nil {
		return "", fmt.Errorf("no parent configured")
	}

	transports := parent.AllTransports()
	if len(transports) == 0 {
		return "", fmt.Errorf("no parent transports configured")
	}

	// Extract peer ID from multiaddr (format: /ip4/.../tcp/.../p2p/PeerID)
	ma, err := multiaddr.NewMultiaddr(transports[0])
	if err != nil {
		return "", fmt.Errorf("invalid parent transport: %w", err)
	}

	peerIDComponent, err := ma.ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		return "", fmt.Errorf("no peer ID in parent transport: %w", err)
	}

	return peer.Decode(peerIDComponent)
}

// ParentTransports returns the parent's multiaddresses
func (n *CoreNode) ParentTransports() []multiaddr.Multiaddr {
	parent := n.Parent()
	if parent == nil {
		return []multiaddr.Multiaddr{}
	}

	transports := parent.AllTransports()
	result := make([]multiaddr.Multiaddr, 0, len(transports))
	
	for _, transport := range transports {
		if ma, err := multiaddr.NewMultiaddr(transport); err == nil {
			result = append(result, ma)
		}
	}
	
	return result
}

// GetTransports returns the transports to reach a given address
func (n *CoreNode) GetTransports(address *OAddress) []multiaddr.Multiaddr {
	// Check if the address already has transports
	leaderTransports := address.LibP2PTransports()
	if len(leaderTransports) > 0 {
		return leaderTransports
	}

	// If no transports provided, search within our network
	if len(leaderTransports) == 0 {
		n.logger.Debug("No leader transports provided, searching within network")
		
		if n.config.Leader == nil {
			if n.Type() == NodeTypeLeader {
				n.logger.Debug("Node is a leader, using own transports")
				transports := n.Transports()
				result := make([]multiaddr.Multiaddr, 0, len(transports))
				for _, transport := range transports {
					if ma, err := multiaddr.NewMultiaddr(transport); err == nil {
						result = append(result, ma)
					}
				}
				return result
			} else {
				n.logger.Warn("Not within a network, cannot search for addressed node without leader")
			}
		} else {
			leaderTransports = n.config.Leader.LibP2PTransports()
		}
	}

	return leaderTransports
}

// HandleStaticAddressTranslation handles translation of static addresses
func (n *CoreNode) HandleStaticAddressTranslation(ctx context.Context, addressInput *OAddress) (*OAddress, error) {
	result := addressInput

	// Handle static address translation
	if !addressInput.HasPrefix("o://leader") {
		// Search for the static address in the leader registry
		searchAddr := NewOAddress("o://leader/register")
		searchAddr.SetTransports(result.LibP2PTransports())

		response, err := n.Use(ctx, searchAddr, "search", map[string]interface{}{
			"staticAddress": result.Root(),
		}, &UseOptions{NoIndex: true})

		if err != nil {
			n.logger.Warnf("Failed to search for static address: %v", err)
			return result, nil
		}

		// Process search results
		if response.Result != nil {
			if searchResults, ok := response.Result.(map[string]interface{})["data"].([]interface{}); ok && len(searchResults) > 0 {
				if firstResult, ok := searchResults[0].(map[string]interface{}); ok {
					if resolvedAddr, ok := firstResult["address"].(string); ok {
						// Add remainder paths to resolved address
						parts := addressInput.SplitAddress()
						if len(parts) > 1 {
							remainderPaths := parts[1:]
							resolvedAddr = fmt.Sprintf("%s/%s", resolvedAddr, strings.Join(remainderPaths, "/"))
						}
						// Convert string slice to interface slice
						transports := make([]interface{}, len(result.AllTransports()))
						for i, t := range result.AllTransports() {
							transports[i] = t
						}
						result = NewOAddress(resolvedAddr, transports...)
					}
				}
			} else {
				n.logger.Warn("Failed to translate static address - no results found")
			}
		}
	}

	return result, nil
}

// TranslateAddress translates an address to determine next hop and target
func (n *CoreNode) TranslateAddress(ctx context.Context, addressWithLeaderTransports *OAddress) (*TranslateAddressResult, error) {
	targetAddress := addressWithLeaderTransports
	
	// Handle static address translation
	var err error
	targetAddress, err = n.HandleStaticAddressTranslation(ctx, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to handle static address translation: %w", err)
	}

	// Resolve the next hop address
	nextHopAddress, err := n.addressResolution.Resolve(ctx, targetAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	// Set transports for the next hop
	leaderTransports := n.GetTransports(nextHopAddress)
	nextHopAddress.SetTransports(leaderTransports)

	return &TranslateAddressResult{
		NextHopAddress: nextHopAddress,
		TargetAddress:  targetAddress,
	}, nil
}

// Use executes a method on a remote address
func (n *CoreNode) Use(ctx context.Context, address *OAddress, method string, params map[string]interface{}, opts *UseOptions) (*OResponse, error) {
	if opts == nil {
		opts = DefaultUseOptions()
	}

	// Translate the address
	result, err := n.TranslateAddress(ctx, address)
	if err != nil {
		n.incrementErrorCount()
		return nil, fmt.Errorf("failed to translate address: %w", err)
	}

	// Connect to the target
	connection, err := n.Connect(ctx, result.NextHopAddress, result.TargetAddress)
	if err != nil {
		n.incrementErrorCount()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer connection.Close()

	// Send the request
	sendParams := &ConnectionSendParams{
		Address: result.TargetAddress.String(),
		Payload: map[string]interface{}{
			"method": method,
			"params": params,
		},
	}

	response, err := connection.Send(ctx, sendParams)
	if err != nil {
		n.incrementErrorCount()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	n.incrementSuccessCount()
	return response, nil
}

// Connect establishes a connection to a target through a next hop
func (n *CoreNode) Connect(ctx context.Context, nextHopAddress, targetAddress *OAddress) (Connection, error) {
	if n.connectionManager == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	params := &ConnectionParams{
		Address:        targetAddress,
		NextHopAddress: nextHopAddress,
		CallerAddress:  n.address,
	}

	connection, err := n.connectionManager.Connect(ctx, params)
	if err != nil {
		if err.Error() == "Can not dial self" {
			return nil, fmt.Errorf("cannot dial self - ensure you're not connecting directly through the leader node")
		}
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	n.logger.Debugf("Successfully connected to: %s", nextHopAddress.String())
	return connection, nil
}

// AdvertiseValueToNetwork advertises a CID to the network
func (n *CoreNode) AdvertiseValueToNetwork(ctx context.Context, value cid.Cid) error {
	if n.p2pNode == nil {
		return fmt.Errorf("p2p node not initialized")
	}

	// For now, we'll simulate the advertisement
	// In a real implementation, this would use the DHT service
	n.logger.Debugf("Advertising CID to network: %s", value.String())
	
	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simulate provide operation (in real implementation, this would use DHT)
	select {
	case <-timeoutCtx.Done():
		return fmt.Errorf("advertise timeout")
	case <-time.After(100 * time.Millisecond): // Simulate network operation
		return nil
	}
}

// AdvertiseToNetwork advertises this node's addresses to the network
func (n *CoreNode) AdvertiseToNetwork(ctx context.Context) error {
	n.logger.Debug("Advertising addresses to network...")

	// Advertise absolute address
	absoluteAddressCid, err := n.address.ToCID()
	if err != nil {
		n.logger.Warnf("Failed to generate CID for absolute address: %v", err)
	} else {
		if err := n.AdvertiseValueToNetwork(ctx, absoluteAddressCid); err != nil {
			n.logger.Warnf("Failed to advertise absolute address: %v", err)
		} else {
			n.logger.Debug("Successfully advertised absolute address")
		}
	}

	// Advertise static address
	staticAddressCid, err := n.staticAddress.ToCID()
	if err != nil {
		n.logger.Warnf("Failed to generate CID for static address: %v", err)
	} else {
		if err := n.AdvertiseValueToNetwork(ctx, staticAddressCid); err != nil {
			n.logger.Warnf("Failed to advertise static address: %v", err)
		} else {
			n.logger.Debug("Successfully advertised static address")
		}
	}

	return nil
}

// Register registers this node with the network leader
func (n *CoreNode) Register(ctx context.Context) error {
	if n.Type() == NodeTypeLeader {
		n.logger.Debug("Skipping registration - node is leader")
		return nil
	}

	n.logger.Debug("Registering node...")

	if n.config.Leader == nil {
		n.logger.Warn("No leader configured, skipping registration")
		return nil
	}

	address := NewOAddress("o://register")
	params := map[string]interface{}{
		"peerId":        n.peerId.String(),
		"address":       n.address.String(),
		"protocols":     []string{}, // Would be populated from p2pNode.GetProtocols()
		"transports":    n.Transports(),
		"staticAddress": n.staticAddress.String(),
	}

	_, err := n.Use(ctx, address, "commit", params, &UseOptions{NoIndex: true})
	if err != nil {
		return fmt.Errorf("failed to register with leader: %w", err)
	}

	n.logger.Debug("Successfully registered with leader")
	return nil
}

// Unregister removes this node from the network
func (n *CoreNode) Unregister(ctx context.Context) error {
	if n.Type() == NodeTypeLeader {
		n.logger.Debug("Skipping unregistration - node is leader")
		return nil
	}

	address := NewOAddress("o://register")
	params := map[string]interface{}{
		"peerId": n.peerId.String(),
	}

	_, err := n.Use(ctx, address, "remove", params, DefaultUseOptions())
	if err != nil {
		n.logger.Warnf("Failed to unregister from network: %v", err)
		return err
	}

	n.logger.Debug("Successfully unregistered from network")
	return nil
}

// incrementSuccessCount increments the success counter (thread-safe)
func (n *CoreNode) incrementSuccessCount() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.successCount++
}

// incrementErrorCount increments the error counter (thread-safe)
func (n *CoreNode) incrementErrorCount() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.errorCount++
}

// Initialize performs node initialization (to be overridden by concrete implementations)
func (n *CoreNode) Initialize(ctx context.Context) error {
	n.logger.Debug("Initializing core node...")
	return nil
}

// Start starts the node
func (n *CoreNode) Start(ctx context.Context) error {
	if n.State() != NodeStateStopped {
		n.logger.Warn("Node is not stopped, skipping start")
		return nil
	}

	n.setState(NodeStateStarting)

	if err := n.Initialize(ctx); err != nil {
		n.setState(NodeStateError)
		n.addError(err)
		return fmt.Errorf("failed to initialize node: %w", err)
	}

	if err := n.Register(ctx); err != nil {
		n.logger.Errorf("Failed to register node: %v", err)
		// Don't fail startup on registration failure
	}

	n.setState(NodeStateRunning)
	n.logger.Info("Node started successfully")
	return nil
}

// Stop stops the node
func (n *CoreNode) Stop(ctx context.Context) error {
	n.logger.Debug("Stopping node...")
	n.setState(NodeStateStopping)

	var errs []error

	// Unregister from network
	if err := n.Unregister(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to unregister: %w", err))
	}

	// Stop libp2p host
	if n.p2pNode != nil {
		if err := n.p2pNode.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close p2p node: %w", err))
		}
	}

	if len(errs) > 0 {
		n.setState(NodeStateError)
		for _, err := range errs {
			n.addError(err)
		}
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	n.setState(NodeStateStopped)
	n.logger.Info("Node stopped successfully")
	return nil
}
