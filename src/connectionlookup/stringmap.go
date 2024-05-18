package connectionlookup

import "sync"

type StringMap struct {
	strings map[string]int
	mu      sync.RWMutex
}

func NewStringMap() *StringMap {
	return &StringMap{strings: map[string]int{}}
}

func (m *StringMap) Get(key string) int {
	m.mu.RLock()
	id, exists := m.strings[key]
	m.mu.RUnlock()
	if exists {
		return id
	}

	m.mu.Lock()
	id = len(m.strings) + 1
	m.strings[key] = id
	m.mu.Unlock()

	return id
}
