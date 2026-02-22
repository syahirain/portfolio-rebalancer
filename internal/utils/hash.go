package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// CanonicalHash computes a SHA256 hash of a map[string]float64 in a canonical way.
// It ensures that the map keys are sorted before hashing to guarantee deterministic output.
func CanonicalHash(allocation map[string]float64) string {
	// Sort keys
	keys := make([]string, 0, len(allocation))
	for k := range allocation {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build a canonical map
	canonical := make(map[string]float64, len(allocation))
	for _, k := range keys {
		canonical[k] = allocation[k]
	}

	// Marshal to JSON
	bytes, _ := json.Marshal(canonical)

	// Compute SHA256
	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", hash)
}
