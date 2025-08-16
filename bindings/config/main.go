// Package main provides CGO bindings for the olane-go config package
package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/olane-labs/olane-go/pkg/config"
	"github.com/olane-labs/olane-go/pkg/utils"
)

// Global storage for libp2p components to keep them alive across C calls
var hosts = make(map[int]host.Host)
var dhts = make(map[int]*dht.IpfsDHT)
var pubsubs = make(map[int]*pubsub.PubSub)
var nextID = 1

// ConfigData represents the configuration data structure for JSON serialization
type ConfigData struct {
	Listeners        []string `json:"listeners"`
	BootstrapPeers   []string `json:"bootstrapPeers"`
	EnableRelay      bool     `json:"enableRelay"`
	EnableDHT        bool     `json:"enableDHT"`
	EnablePubsub     bool     `json:"enablePubsub"`
	DHTProtocolPrefix string  `json:"dhtProtocolPrefix"`
	KBucketSize      int      `json:"kBucketSize"`
}

// NodeInfo represents information about a created node
type NodeInfo struct {
	ID         int      `json:"id"`
	PeerID     string   `json:"peerId"`
	Addrs      []string `json:"addrs"`
	HasDHT     bool     `json:"hasDHT"`
	HasPubsub  bool     `json:"hasPubsub"`
	Protocols  []string `json:"protocols"`
}

//export get_default_config
func get_default_config() *C.char {
	cfg := config.DefaultLibp2pConfig()
	
	configData := ConfigData{
		Listeners:         cfg.Listeners,
		BootstrapPeers:    cfg.BootstrapPeers,
		EnableRelay:       cfg.EnableRelay,
		EnableDHT:         cfg.EnableDHT,
		EnablePubsub:      cfg.EnablePubsub,
		DHTProtocolPrefix: string(cfg.DHTProtocolPrefix),
		KBucketSize:       cfg.KBucketSize,
	}
	
	jsonData, err := json.Marshal(configData)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to marshal config: %v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export create_config
func create_config(listenersJson *C.char, bootstrapPeersJson *C.char, enableRelay C.int, enableDHT C.int, enablePubsub C.int, kBucketSize C.int) *C.char {
	// Parse listeners
	var listeners []string
	if listenersJson != nil {
		listenersStr := C.GoString(listenersJson)
		if err := json.Unmarshal([]byte(listenersStr), &listeners); err != nil {
			return C.CString(fmt.Sprintf(`{"error": "invalid listeners JSON: %v"}`, err))
		}
	}
	
	// Parse bootstrap peers
	var bootstrapPeers []string
	if bootstrapPeersJson != nil {
		bootstrapStr := C.GoString(bootstrapPeersJson)
		if err := json.Unmarshal([]byte(bootstrapStr), &bootstrapPeers); err != nil {
			return C.CString(fmt.Sprintf(`{"error": "invalid bootstrap peers JSON: %v"}`, err))
		}
	}
	
	// Create configuration
	cfg := config.DefaultLibp2pConfig()
	
	if len(listeners) > 0 {
		cfg.Listeners = listeners
	}
	if len(bootstrapPeers) > 0 {
		cfg.BootstrapPeers = bootstrapPeers
	}
	
	cfg.EnableRelay = enableRelay != 0
	cfg.EnableDHT = enableDHT != 0
	cfg.EnablePubsub = enablePubsub != 0
	
	if kBucketSize > 0 {
		cfg.KBucketSize = int(kBucketSize)
	}
	
	// Return the configuration as JSON
	configData := ConfigData{
		Listeners:         cfg.Listeners,
		BootstrapPeers:    cfg.BootstrapPeers,
		EnableRelay:       cfg.EnableRelay,
		EnableDHT:         cfg.EnableDHT,
		EnablePubsub:      cfg.EnablePubsub,
		DHTProtocolPrefix: string(cfg.DHTProtocolPrefix),
		KBucketSize:       cfg.KBucketSize,
	}
	
	jsonData, err := json.Marshal(configData)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to marshal config: %v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export create_node
func create_node(configJson *C.char) *C.char {
	configStr := C.GoString(configJson)
	
	var configData ConfigData
	if err := json.Unmarshal([]byte(configStr), &configData); err != nil {
		return C.CString(fmt.Sprintf(`{"error": "invalid config JSON: %v"}`, err))
	}
	
	// Create libp2p configuration
	cfg := config.DefaultLibp2pConfig()
	cfg.Listeners = configData.Listeners
	cfg.BootstrapPeers = configData.BootstrapPeers
	cfg.EnableRelay = configData.EnableRelay
	cfg.EnableDHT = configData.EnableDHT
	cfg.EnablePubsub = configData.EnablePubsub
	cfg.KBucketSize = configData.KBucketSize
	
	// Create the node with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	host, kadDHT, gossipSub, err := config.CreateNode(ctx, cfg)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to create node: %v"}`, err))
	}
	
	// Store components with unique ID
	nodeID := nextID
	nextID++
	
	hosts[nodeID] = host
	if kadDHT != nil {
		dhts[nodeID] = kadDHT
	}
	if gossipSub != nil {
		pubsubs[nodeID] = gossipSub
	}
	
	// Get node information
	addrs := make([]string, len(host.Addrs()))
	for i, addr := range host.Addrs() {
		addrs[i] = fmt.Sprintf("%s/p2p/%s", addr.String(), host.ID().String())
	}
	
	protocols := host.Mux().Protocols()
	protocolStrs := make([]string, len(protocols))
	for i, p := range protocols {
		protocolStrs[i] = string(p)
	}
	
	nodeInfo := NodeInfo{
		ID:        nodeID,
		PeerID:    host.ID().String(),
		Addrs:     addrs,
		HasDHT:    kadDHT != nil,
		HasPubsub: gossipSub != nil,
		Protocols: protocolStrs,
	}
	
	jsonData, err := json.Marshal(nodeInfo)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to marshal node info: %v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export connect_to_bootstrap_peers
func connect_to_bootstrap_peers(nodeID C.int, bootstrapPeersJson *C.char) *C.char {
	host, exists := hosts[int(nodeID)]
	if !exists {
		return C.CString(`{"error": "node not found"}`)
	}
	
	var bootstrapPeers []string
	if bootstrapPeersJson != nil {
		bootstrapStr := C.GoString(bootstrapPeersJson)
		if err := json.Unmarshal([]byte(bootstrapStr), &bootstrapPeers); err != nil {
			return C.CString(fmt.Sprintf(`{"error": "invalid bootstrap peers JSON: %v"}`, err))
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := config.ConnectToBootstrapPeers(ctx, host, bootstrapPeers); err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to connect to bootstrap peers: %v"}`, err))
	}
	
	return C.CString(`{"success": true}`)
}

//export get_node_info
func get_node_info(nodeID C.int) *C.char {
	host, exists := hosts[int(nodeID)]
	if !exists {
		return C.CString(`{"error": "node not found"}`)
	}
	
	addrs := make([]string, len(host.Addrs()))
	for i, addr := range host.Addrs() {
		addrs[i] = fmt.Sprintf("%s/p2p/%s", addr.String(), host.ID().String())
	}
	
	protocols := host.Mux().Protocols()
	protocolStrs := make([]string, len(protocols))
	for i, p := range protocols {
		protocolStrs[i] = string(p)
	}
	
	// Get connected peers
	peers := host.Network().Peers()
	peerStrs := make([]string, len(peers))
	for i, p := range peers {
		peerStrs[i] = p.String()
	}
	
	nodeInfo := map[string]interface{}{
		"id":             int(nodeID),
		"peerId":         host.ID().String(),
		"addrs":          addrs,
		"hasDHT":         dhts[int(nodeID)] != nil,
		"hasPubsub":      pubsubs[int(nodeID)] != nil,
		"protocols":      protocolStrs,
		"connectedPeers": peerStrs,
		"peerCount":      len(peers),
	}
	
	jsonData, err := json.Marshal(nodeInfo)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to marshal node info: %v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export close_node
func close_node(nodeID C.int) *C.char {
	nodeIDInt := int(nodeID)
	
	// Close host
	if host, exists := hosts[nodeIDInt]; exists {
		if err := host.Close(); err != nil {
			return C.CString(fmt.Sprintf(`{"error": "failed to close host: %v"}`, err))
		}
		delete(hosts, nodeIDInt)
	}
	
	// Close DHT
	if kadDHT, exists := dhts[nodeIDInt]; exists {
		if err := kadDHT.Close(); err != nil {
			return C.CString(fmt.Sprintf(`{"error": "failed to close DHT: %v"}`, err))
		}
		delete(dhts, nodeIDInt)
	}
	
	// Remove pubsub (it doesn't have a Close method)
	if _, exists := pubsubs[nodeIDInt]; exists {
		delete(pubsubs, nodeIDInt)
	}
	
	return C.CString(`{"success": true}`)
}

//export get_peer_count
func get_peer_count(nodeID C.int) C.int {
	host, exists := hosts[int(nodeID)]
	if !exists {
		return -1
	}
	
	return C.int(len(host.Network().Peers()))
}

//export get_connected_peers
func get_connected_peers(nodeID C.int) *C.char {
	host, exists := hosts[int(nodeID)]
	if !exists {
		return C.CString(`{"error": "node not found"}`)
	}
	
	peers := host.Network().Peers()
	peerStrs := make([]string, len(peers))
	for i, p := range peers {
		peerStrs[i] = p.String()
	}
	
	result := map[string]interface{}{
		"peers": peerStrs,
		"count": len(peers),
	}
	
	jsonData, err := json.Marshal(result)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to marshal peers: %v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export validate_multiaddr
func validate_multiaddr(addr *C.char) C.int {
	addrStr := C.GoString(addr)
	_, err := utils.ParseMultiaddr(addrStr)
	if err != nil {
		return 0 // false
	}
	return 1 // true
}

//export validate_multiaddrs
func validate_multiaddrs(addrsJson *C.char) *C.char {
	addrsStr := C.GoString(addrsJson)
	
	var addrs []string
	if err := json.Unmarshal([]byte(addrsStr), &addrs); err != nil {
		return C.CString(fmt.Sprintf(`{"error": "invalid addresses JSON: %v"}`, err))
	}
	
	if err := utils.ValidateMultiaddrs(addrs); err != nil {
		return C.CString(fmt.Sprintf(`{"error": "validation failed: %v"}`, err))
	}
	
	return C.CString(`{"valid": true}`)
}

// Note: We don't provide a free_string function to avoid memory management issues.
// The Go runtime will handle cleanup of strings returned by C.CString().
// This may cause minor memory leaks but prevents crashes.

// Required main function for CGO
func main() {}
