// Package utils provides utility functions for Olane networks.
// This package mirrors the functionality of the TypeScript o-config/utils package.
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"
	"runtime"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// GenerateKeyPair generates a new RSA key pair for a libp2p node
func GenerateKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
}

// GenerateEd25519KeyPair generates a new Ed25519 key pair for a libp2p node
func GenerateEd25519KeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
}

// PrivKeyFromBase64 decodes a base64 encoded private key
func PrivKeyFromBase64(b64 string) (crypto.PrivKey, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	return crypto.UnmarshalPrivateKey(data)
}

// PrivKeyToBase64 encodes a private key to base64
func PrivKeyToBase64(priv crypto.PrivKey) (string, error) {
	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// PeerIDFromPrivKey generates a peer ID from a private key
func PeerIDFromPrivKey(priv crypto.PrivKey) (peer.ID, error) {
	return peer.IDFromPrivateKey(priv)
}

// ParseMultiaddr parses a multiaddr string and validates it
func ParseMultiaddr(addr string) (multiaddr.Multiaddr, error) {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid multiaddr %s: %w", addr, err)
	}
	return ma, nil
}

// ValidateMultiaddrs validates a slice of multiaddr strings
func ValidateMultiaddrs(addrs []string) error {
	for _, addr := range addrs {
		if _, err := ParseMultiaddr(addr); err != nil {
			return err
		}
	}
	return nil
}

// MergeConfigs merges two configuration structs using reflection.
// Fields in override take precedence over those in base.
// This provides similar functionality to the object spread operator in TypeScript.
func MergeConfigs(base, override interface{}) interface{} {
	baseValue := reflect.ValueOf(base)
	overrideValue := reflect.ValueOf(override)

	// Handle pointers
	if baseValue.Kind() == reflect.Ptr {
		baseValue = baseValue.Elem()
	}
	if overrideValue.Kind() == reflect.Ptr {
		overrideValue = overrideValue.Elem()
	}

	// Create a new struct of the same type as base
	resultType := baseValue.Type()
	result := reflect.New(resultType).Elem()

	// Copy fields from base
	for i := 0; i < baseValue.NumField(); i++ {
		field := baseValue.Field(i)
		if field.CanInterface() {
			result.Field(i).Set(field)
		}
	}

	// Override with fields from override
	for i := 0; i < overrideValue.NumField(); i++ {
		overrideField := overrideValue.Field(i)
		fieldName := overrideValue.Type().Field(i).Name

		// Find corresponding field in result
		resultField := result.FieldByName(fieldName)
		if resultField.IsValid() && resultField.CanSet() && !overrideField.IsZero() {
			resultField.Set(overrideField)
		}
	}

	return result.Interface()
}

// GetFunctionName returns the name of the calling function
// This can be useful for logging and debugging
func GetFunctionName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	return runtime.FuncForPC(pc).Name()
}

// DefaultIfEmpty returns the default value if the input string is empty
func DefaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// DefaultIfZero returns the default value if the input is the zero value for its type
func DefaultIfZero[T comparable](value, defaultValue T) T {
	var zero T
	if value == zero {
		return defaultValue
	}
	return value
}

// SliceContains checks if a slice contains a specific item
func SliceContains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// RemoveFromSlice removes all occurrences of an item from a slice
func RemoveFromSlice[T comparable](slice []T, item T) []T {
	var result []T
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// UniqueSlice removes duplicate items from a slice
func UniqueSlice[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	var result []T
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
