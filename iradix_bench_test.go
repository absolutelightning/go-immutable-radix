package iradix

import (
	"fmt"
	"testing"
)

// generateKeysForDenseNode generates all 256 keys for a given prefix.
func generateKeysForDenseNode(prefix string) []string {
	var keys []string
	for i := 0; i < 256; i++ {
		// Append characters within the range 'a' to 'z' for readable test data
		keys = append(keys, fmt.Sprintf("%s%c", prefix, 'a'+(i%26)))
	}
	return keys
}

// generateDenseTestData generates an array of strings representing dense inputs.
// Each node will have 256 edges, and the depth of the tree is controlled by levels.
func generateDenseTestData(levels int) []string {
	var inputs []string
	var prefixes []string
	prefixes = append(prefixes, "") // Start with the root prefix

	for level := 0; level < levels; level++ {
		var newPrefixes []string
		for _, prefix := range prefixes {
			// Generate 256 edges for each prefix
			keys := generateKeysForDenseNode(prefix)
			inputs = append(inputs, keys...)           // Add all keys to the input list
			newPrefixes = append(newPrefixes, keys...) // Use these keys as prefixes for the next level
		}
		prefixes = newPrefixes
	}
	return inputs
}

func BenchmarkDenseTreeMemoryInsert(b *testing.B) {
	tr := New()

	strs := generateDenseTestData(3)

	b.ResetTimer()

	for _, str := range strs {
		tr, _, _ = tr.Insert([]byte(str), nil)
	}
}

func BenchmarkDenseTreeMemorySearch(b *testing.B) {
	tr := New()

	strs := generateDenseTestData(3)

	for _, str := range strs {
		tr, _, _ = tr.Insert([]byte(str), nil)
	}

	b.ResetTimer()

	for _, str := range strs {
		tr.Get([]byte(str))
	}

}
