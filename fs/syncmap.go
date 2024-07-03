package fs

import "sync"

type SyncMap[K any, V any] struct {
	m sync.Map
}

func NewSyncMap[K, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

func (sm *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := sm.m.Load(key)
	if !ok {
		return
	}

	value, ok = v.(V)
	return
}

func (sm *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := sm.m.LoadOrStore(key, value)
	if !loaded {
		actual = value
	} else {
		actual = v.(V)
	}

	return
}

func (sm *SyncMap[K, V]) Delete(key K) {
	sm.m.Delete(key)
}

func (sm *SyncMap[K, V]) Len() int {
	count := 0
	sm.m.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (sm *SyncMap[K, V]) Keys() []K {
	keys := make([]K, 0, sm.Len())
	sm.m.Range(func(key, _ any) bool {
		keys = append(keys, key.(K))
		return true
	})
	return keys
}
