// Package main provides CGO bindings for Python integration
package main

import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/olane-labs/olane-go/pkg/config"
	"github.com/olane-labs/olane-go/pkg/core"
)

// Global storage for nodes to keep them alive across C calls
var nodes = make(map[int]*core.CoreNode)
var nextNodeID = 1

//export create_node
func create_node(addressC *C.char, nodeTypeC *C.char, nameC *C.char, descriptionC *C.char) C.int {
	address := C.GoString(addressC)
	nodeType := C.GoString(nodeTypeC)
	name := C.GoString(nameC)
	description := C.GoString(descriptionC)

	cfg := core.DefaultCoreConfig()
	cfg.Address = core.NewOAddress(address)
	cfg.Name = name
	cfg.Description = description
	
	switch nodeType {
	case "leader":
		cfg.Type = core.NodeTypeLeader
	case "node":
		cfg.Type = core.NodeTypeNode
	case "tool":
		cfg.Type = core.NodeTypeTool
	case "agent":
		cfg.Type = core.NodeTypeAgent
	default:
		cfg.Type = core.NodeTypeUnknown
	}

	node := core.NewCoreNode(cfg)
	
	nodeID := nextNodeID
	nextNodeID++
	nodes[nodeID] = node
	
	return C.int(nodeID)
}

//export start_node
func start_node(nodeID C.int) *C.char {
	node, exists := nodes[int(nodeID)]
	if !exists {
		return C.CString("error: node not found")
	}

	ctx := context.Background()
	if err := node.Start(ctx); err != nil {
		return C.CString(fmt.Sprintf("error: %v", err))
	}

	return C.CString("success")
}

//export stop_node
func stop_node(nodeID C.int) *C.char {
	node, exists := nodes[int(nodeID)]
	if !exists {
		return C.CString("error: node not found")
	}

	ctx := context.Background()
	if err := node.Stop(ctx); err != nil {
		return C.CString(fmt.Sprintf("error: %v", err))
	}

	return C.CString("success")
}

//export node_whoami
func node_whoami(nodeID C.int) *C.char {
	node, exists := nodes[int(nodeID)]
	if !exists {
		return C.CString(`{"error": "node not found"}`)
	}

	ctx := context.Background()
	whoami, err := node.WhoAmI(ctx)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "%v"}`, err))
	}

	jsonData, err := json.Marshal(whoami)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "json marshal failed: %v"}`, err))
	}

	return C.CString(string(jsonData))
}

//export create_address
func create_address(addressC *C.char) C.int {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	
	// For simplicity, we'll return the hash code as ID
	// In production, you'd want a proper registry
	return C.int(len(address))
}

//export address_validate
func address_validate(addressC *C.char) C.int {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	
	if addr.Validate() {
		return 1
	}
	return 0
}

//export address_get_root
func address_get_root(addressC *C.char) *C.char {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	return C.CString(addr.Root())
}

//export address_get_paths
func address_get_paths(addressC *C.char) *C.char {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	return C.CString(addr.Paths())
}

//export address_is_leader
func address_is_leader(addressC *C.char) C.int {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	
	if addr.IsLeaderAddress() {
		return 1
	}
	return 0
}

//export address_is_tool
func address_is_tool(addressC *C.char) C.int {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	
	if addr.IsToolAddress() {
		return 1
	}
	return 0
}

//export address_get_cid
func address_get_cid(addressC *C.char) *C.char {
	address := C.GoString(addressC)
	addr := core.NewOAddress(address)
	
	cid, err := addr.ToCID()
	if err != nil {
		return C.CString(fmt.Sprintf("error: %v", err))
	}
	
	return C.CString(cid.String())
}

//export create_libp2p_config
func create_libp2p_config() *C.char {
	cfg := config.DefaultLibp2pConfig()
	
	result := map[string]interface{}{
		"listeners":     cfg.Listeners,
		"enableDHT":     cfg.EnableDHT,
		"enablePubsub":  cfg.EnablePubsub,
		"enableRelay":   cfg.EnableRelay,
		"kBucketSize":   cfg.KBucketSize,
	}
	
	jsonData, err := json.Marshal(result)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "%v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export create_libp2p_node
func create_libp2p_node(listenersC *C.char) *C.char {
	listenersJson := C.GoString(listenersC)
	
	var listeners []string
	if err := json.Unmarshal([]byte(listenersJson), &listeners); err != nil {
		return C.CString(fmt.Sprintf(`{"error": "invalid listeners json: %v"}`, err))
	}
	
	cfg := config.DefaultLibp2pConfig()
	cfg.Listeners = listeners
	
	ctx := context.Background()
	host, dht, pubsub, err := config.CreateNode(ctx, cfg)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "failed to create node: %v"}`, err))
	}
	
	result := map[string]interface{}{
		"peerId":    host.ID().String(),
		"addrs":     []string{},
		"hasDHT":    dht != nil,
		"hasPubsub": pubsub != nil,
	}
	
	for _, addr := range host.Addrs() {
		result["addrs"] = append(result["addrs"].([]string), addr.String())
	}
	
	jsonData, err := json.Marshal(result)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "%v"}`, err))
	}
	
	return C.CString(string(jsonData))
}

//export free_string
func free_string(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export cleanup_node
func cleanup_node(nodeID C.int) {
	delete(nodes, int(nodeID))
}

// Required main function for CGO
func main() {}
