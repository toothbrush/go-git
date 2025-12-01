package packfile

import (
	"sort"

	"github.com/go-git/go-git/v6/plumbing"
)

// offsetEntry stores the mapping from packfile offset to object hash.
type offsetEntry struct {
	Offset int64
	Hash   plumbing.Hash
}

// offsetIndex provides a memory-efficient mapping from packfile offsets to
// object hashes. It uses a sorted slice instead of a map to reduce memory
// overhead, since offsets are typically added in order during parsing.
//
// For a 10M object packfile:
//   - Full ObjectHeader cache: ~1.2GB
//   - offsetIndex: ~280MB (28 bytes per entry)
type offsetIndex struct {
	entries []offsetEntry
	sorted  bool
}

// newOffsetIndex creates a new offset index with the given capacity hint.
func newOffsetIndex(capacity int) *offsetIndex {
	return &offsetIndex{
		entries: make([]offsetEntry, 0, capacity),
		sorted:  true,
	}
}

// Add adds an offset→hash mapping to the index.
func (idx *offsetIndex) Add(offset int64, hash plumbing.Hash) {
	// Check if we need to mark as unsorted
	if len(idx.entries) > 0 && offset < idx.entries[len(idx.entries)-1].Offset {
		idx.sorted = false
	}
	idx.entries = append(idx.entries, offsetEntry{Offset: offset, Hash: hash})
}

// Lookup returns the hash for the given offset, or false if not found.
func (idx *offsetIndex) Lookup(offset int64) (plumbing.Hash, bool) {
	if len(idx.entries) == 0 {
		return plumbing.ZeroHash, false
	}

	// Ensure sorted for binary search
	if !idx.sorted {
		sort.Slice(idx.entries, func(i, j int) bool {
			return idx.entries[i].Offset < idx.entries[j].Offset
		})
		idx.sorted = true
	}

	// Binary search
	i := sort.Search(len(idx.entries), func(i int) bool {
		return idx.entries[i].Offset >= offset
	})

	if i < len(idx.entries) && idx.entries[i].Offset == offset {
		return idx.entries[i].Hash, true
	}

	return plumbing.ZeroHash, false
}

// Len returns the number of entries in the index.
func (idx *offsetIndex) Len() int {
	return len(idx.entries)
}
