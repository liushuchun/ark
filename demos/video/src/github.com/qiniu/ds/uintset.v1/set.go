package uintset

// ---------------------------------------------------------------------------

type empty struct{}

// Type is a set of uint, implemented via map[uint]struct{} for minimal memory consumption.
type Type map[uint]empty

// New creates a Type from a list of values.
func New(items ...uint) Type {
	ss := Type{}
	ss.Insert(items...)
	return ss
}

// Add adds one item to the set.
func (s Type) Add(item uint) {
	s[item] = empty{}
}

// Insert adds items to the set.
func (s Type) Insert(items ...uint) {
	for _, item := range items {
		s[item] = empty{}
	}
}

// Delete removes all items from the set.
func (s Type) Delete(items ...uint) {
	for _, item := range items {
		delete(s, item)
	}
}

// Has returns true iff item is contained in the set.
func (s Type) Has(item uint) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true iff all items are contained in the set.
func (s Type) HasAll(items ...uint) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// IsSuperset returns true iff s1 is a superset of s2.
func (s Type) IsSuperset(s2 Type) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Len returns the size of the set.
func (s Type) Len() int {
	return len(s)
}

// ---------------------------------------------------------------------------
