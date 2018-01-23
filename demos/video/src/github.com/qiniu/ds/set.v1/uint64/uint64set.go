package set

// ---------------------------------------------------------------------------

type empty struct{}

// Uint64Set is a set of strings, implemented via map[uint64]struct{} for minimal memory consumption.
type Uint64Set map[uint64]empty

// NewUint64Set creates a Uint64Set from a list of values.
func NewUint64Set(items ...uint64) Uint64Set {
	ss := Uint64Set{}
	ss.Insert(items...)
	return ss
}

// Add adds one item to the set.
func (s Uint64Set) Add(item uint64) {
	s[item] = empty{}
}

// Insert adds items to the set.
func (s Uint64Set) Insert(items ...uint64) {
	for _, item := range items {
		s[item] = empty{}
	}
}

// Delete removes all items from the set.
func (s Uint64Set) Delete(items ...uint64) {
	for _, item := range items {
		delete(s, item)
	}
}

// Has returns true iff item is contained in the set.
func (s Uint64Set) Has(item uint64) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true iff all items are contained in the set.
func (s Uint64Set) HasAll(items ...uint64) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// IsSuperset returns true iff s1 is a superset of s2.
func (s1 Uint64Set) IsSuperset(s2 Uint64Set) bool {
	for item := range s2 {
		if !s1.Has(item) {
			return false
		}
	}
	return true
}

// Len returns the size of the set.
func (s Uint64Set) Len() int {
	return len(s)
}

// ---------------------------------------------------------------------------

