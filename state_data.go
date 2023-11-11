package goap

import (
	"math"
	"sort"
	"sync"
)

const (
	linearCutoff = 0 // 1 cache line
)

type elem uint64

func elemOf(f fact, e expr) elem {
	return elem(f)<<32 | elem(e)
}

func (e elem) Key() fact {
	return fact(e >> 32)
}

func (e elem) Value() expr {
	return expr(e & 0xFFFFFFFF)
}

// hashset is a map-like data-structure for state
type hashset struct {
	data []elem // Keys and values, interleaved
}

var pool = sync.Pool{
	New: func() interface{} {
		return &hashset{
			data: make([]elem, 0, 16),
		}
	},
}

// newHashSet returns a map initialized with n spaces.
func newHashSet(size int) *hashset {
	if size <= 0 {
		size = 2
	}

	m := pool.Get().(*hashset)

	capacity := arraySize(size, 0.9)
	if cap(m.data) < capacity {
		m.data = make([]elem, 0, capacity)
		return m
	}

	m.data = m.data[:0]
	return m
}

func (m *hashset) Release() {
	for i := range m.data {
		m.data[i] = 0
	}

	m.data = m.data[:0]
	pool.Put(m)
}

func (m *hashset) findLinear(key fact) (int, bool) {
	for i := 0; i < len(m.data); i++ {
		if key == m.data[i].Key() {
			return i, true
		}
	}
	return 0, false
}

func (m *hashset) find(key fact) (int, bool) {
	if m.Count() <= linearCutoff {
		return m.findLinear(key)
	}

	x := sort.Search(len(m.data), func(i int) bool { return m.data[i].Key() >= key })
	if x < len(m.data) && m.data[x].Key() == key {
		return x, true
	}
	return x, false
}

func (m *hashset) sort() {
	if m.Count() > linearCutoff {
		sort.Slice(m.data, func(i, j int) bool { return m.data[i].Key() < m.data[j].Key() })
	}
}

// Load returns the value stored in the map for a key, or nil if no value is
// present. The ok result indicates whether value was found in the map.
func (m *hashset) Load(key fact) (expr, bool) {
	if i, ok := m.find(key); ok {
		return m.data[i].Value(), true
	}
	return 0, false
}

// Store sets the value for a key.
func (m *hashset) Store(f fact, e expr) {
	if i, ok := m.find(f); ok {
		m.data[i] = elemOf(f, e)
		return
	}

	m.data = append(m.data, elemOf(f, e))
	m.sort()
}

// Delete deletes the value for a key.
func (m *hashset) Delete(f fact) {
	i, ok := m.find(f)
	if !ok {
		return
	}

	m.data[i] = 0
	m.sort()
}

// Count returns number of key/value pairs in the map.
func (m *hashset) Count() int {
	return len(m.data)
}

// Range calls f sequentially for each key and value present in the map.
func (m *hashset) Range(fn func(f fact, e expr)) {
	for _, v := range m.data {
		fn(v.Key(), v.Value())
	}
}

// RangeErr calls f sequentially for each key and value present in the map. If fn
// returns error, range stops the iteration.
func (m *hashset) RangeErr(fn func(f fact, e expr) error) error {
	for _, v := range m.data {
		if err := fn(v.Key(), v.Value()); err != nil {
			return err
		}
	}
	return nil
}

// Clone returns a copy of the map.
func (m *hashset) Clone() *hashset {
	clone := newHashSet(len(m.data))
	clone.data = clone.data[:len(m.data)]
	copy(clone.data, m.data)
	return clone
}

func capacityFor(x uint32) uint32 {
	if x == math.MaxUint32 {
		return x
	}

	if x == 0 {
		return 1
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

func arraySize(size int, fill float64) int {
	s := capacityFor(uint32(math.Ceil(float64(size) / fill)))
	if s < 2 {
		s = 2
	}

	return int(s)
}
