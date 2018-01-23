package set

// ---------------------------------------------------------------------------

type empty struct{}

// UintSet is a set of strings, implemented via map[uint]struct{} for minimal memory consumption.
type UintSet map[uint]empty

// NewUintSet creates a UintSet from a list of values.
func NewUintSet(items ...uint) UintSet {
	ss := UintSet{}
	ss.Insert(items...)
	return ss
}

// Add adds one item to the set.
func (s UintSet) Add(item uint) {
	s[item] = empty{}
}

// Insert adds items to the set.
func (s UintSet) Insert(items ...uint) {
	for _, item := range items {
		s[item] = empty{}
	}
}

// Delete removes all items from the set.
func (s UintSet) Delete(items ...uint) {
	for _, item := range items {
		delete(s, item)
	}
}

// Has returns true iff item is contained in the set.
func (s UintSet) Has(item uint) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true iff all items are contained in the set.
func (s UintSet) HasAll(items ...uint) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// IsSuperset returns true iff s1 is a superset of s2.
func (s1 UintSet) IsSuperset(s2 UintSet) bool {
	for item := range s2 {
		if !s1.Has(item) {
			return false
		}
	}
	return true
}

// Len returns the size of the set.
func (s UintSet) Len() int {
	return len(s)
}

// ---------------------------------------------------------------------------

