package core

import (
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

// OAddress represents an o-protocol address (o://)
type OAddress struct {
	value      string
	transports []interface{} // Can be multiaddr.Multiaddr or string
}

// NewOAddress creates a new OAddress
func NewOAddress(value string, transports ...interface{}) *OAddress {
	return &OAddress{
		value:      value,
		transports: transports,
	}
}

// Value returns the string value of the address
func (addr *OAddress) Value() string {
	return addr.value
}

// SetTransports sets the transports for this address
func (addr *OAddress) SetTransports(transports []multiaddr.Multiaddr) {
	addr.transports = make([]interface{}, len(transports))
	for i, t := range transports {
		addr.transports[i] = t
	}
}

// SetTransportsFromStrings sets the transports from string values
func (addr *OAddress) SetTransportsFromStrings(transports []string) {
	addr.transports = make([]interface{}, len(transports))
	for i, t := range transports {
		addr.transports[i] = t
	}
}

// LibP2PTransports returns only the multiaddr.Multiaddr transports
func (addr *OAddress) LibP2PTransports() []multiaddr.Multiaddr {
	var result []multiaddr.Multiaddr
	for _, t := range addr.transports {
		if ma, ok := t.(multiaddr.Multiaddr); ok {
			result = append(result, ma)
		}
	}
	return result
}

// CustomTransports returns only the string transports
func (addr *OAddress) CustomTransports() []string {
	var result []string
	for _, t := range addr.transports {
		if s, ok := t.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// AllTransports returns all transports as strings
func (addr *OAddress) AllTransports() []string {
	var result []string
	for _, t := range addr.transports {
		switch v := t.(type) {
		case multiaddr.Multiaddr:
			result = append(result, v.String())
		case string:
			result = append(result, v)
		}
	}
	return result
}

// Validate checks if the address is valid
func (addr *OAddress) Validate() bool {
	return strings.HasPrefix(addr.value, "o://")
}

// Paths returns the path portion of the address (without o://)
func (addr *OAddress) Paths() string {
	return strings.TrimPrefix(addr.value, "o://")
}

// Protocol returns the protocol form of the address (/o/...)
func (addr *OAddress) Protocol() string {
	return strings.Replace(addr.value, "o://", "/o/", 1)
}

// Root returns the root portion of the address
func (addr *OAddress) Root() string {
	paths := addr.Paths()
	if paths == "" {
		return addr.value
	}
	
	parts := strings.Split(paths, "/")
	if len(parts) == 0 {
		return addr.value
	}
	
	return "o://" + parts[0]
}

// String returns the string representation of the address
func (addr *OAddress) String() string {
	return addr.value
}

// ToMultiaddr converts the address to a multiaddr format
func (addr *OAddress) ToMultiaddr() (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(addr.Protocol())
}

// FromMultiaddr creates an OAddress from a multiaddr
func FromMultiaddr(ma multiaddr.Multiaddr) *OAddress {
	value := strings.Replace(ma.String(), "/o/", "o://", 1)
	return NewOAddress(value)
}

// Equals checks if two addresses are equal
func (addr *OAddress) Equals(other *OAddress) bool {
	return addr.value == other.value
}

// ToCID generates a CID for this address
func (addr *OAddress) ToCID() (cid.Cid, error) {
	// Convert to JSON-like bytes
	jsonBytes := []byte(fmt.Sprintf(`{"address":"%s"}`, addr.String()))
	
	// Create a multihash
	mh, err := multihash.Sum(jsonBytes, multihash.SHA2_256, -1)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create multihash: %w", err)
	}
	
	// Create CID v1 with DAG-JSON codec
	c := cid.NewCidV1(cid.DagJSON, mh)
	return c, nil
}

// ParseOAddress parses a string into an OAddress
func ParseOAddress(value string) (*OAddress, error) {
	if !strings.HasPrefix(value, "o://") {
		return nil, fmt.Errorf("invalid o-address: must start with o://")
	}
	
	return NewOAddress(value), nil
}

// ChildAddress creates a child address under a parent
func ChildAddress(parent, child *OAddress) *OAddress {
	parentPath := parent.Paths()
	childPath := child.Paths()
	
	// If parent already has a path, append the child
	if parentPath != "" {
		return NewOAddress(fmt.Sprintf("o://%s/%s", parentPath, childPath))
	}
	
	// If parent is just a root, add the child path
	return NewOAddress(fmt.Sprintf("%s/%s", parent.String(), childPath))
}

// SplitAddress splits an address into its components
func (addr *OAddress) SplitAddress() []string {
	paths := addr.Paths()
	if paths == "" {
		return []string{}
	}
	return strings.Split(paths, "/")
}

// HasPrefix checks if the address starts with the given prefix
func (addr *OAddress) HasPrefix(prefix string) bool {
	return strings.HasPrefix(addr.value, prefix)
}

// WithPath appends a path to the address
func (addr *OAddress) WithPath(path string) *OAddress {
	newValue := addr.value
	if !strings.HasSuffix(newValue, "/") && !strings.HasPrefix(path, "/") {
		newValue += "/"
	}
	newValue += path
	return NewOAddress(newValue, addr.transports...)
}

// WithTransports creates a copy of the address with new transports
func (addr *OAddress) WithTransports(transports ...interface{}) *OAddress {
	return NewOAddress(addr.value, transports...)
}

// Clone creates a copy of the address
func (addr *OAddress) Clone() *OAddress {
	newAddr := &OAddress{
		value:      addr.value,
		transports: make([]interface{}, len(addr.transports)),
	}
	copy(newAddr.transports, addr.transports)
	return newAddr
}

// IsLeaderAddress checks if this is a leader address
func (addr *OAddress) IsLeaderAddress() bool {
	return addr.HasPrefix("o://leader")
}

// IsToolAddress checks if this address points to a tool
func (addr *OAddress) IsToolAddress() bool {
	parts := addr.SplitAddress()
	return len(parts) > 1 && (parts[1] == "tool" || parts[1] == "tools")
}

// GetToolName extracts the tool name from a tool address
func (addr *OAddress) GetToolName() string {
	if !addr.IsToolAddress() {
		return ""
	}
	
	parts := addr.SplitAddress()
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// GetMethod extracts the method from an address
func (addr *OAddress) GetMethod() string {
	parts := addr.SplitAddress()
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
