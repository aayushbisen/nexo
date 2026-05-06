package store

import "sync"

type Store[V any] struct {
	mut  sync.RWMutex
	data map[string]V
}

func New[V any]() *Store[V] {

	d := make(map[string]V)

	return &Store[V]{
		data: d,
	}

}

func (s *Store[V]) Set(key string, value V) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.data[key] = value

}
func (s *Store[V]) Get(key string) (V, bool) {
	var zero V
	s.mut.RLock()
	defer s.mut.RUnlock()

	if val, ok := s.data[key]; ok {
		return val, ok
	}
	return zero, false
}

func (s *Store[V]) Delete(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	delete(s.data, key)
}
