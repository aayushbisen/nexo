package store

import (
	"container/list"
	"sync"
)

type entry[V any] struct {
	key   string
	value V
}

type Store[V any] struct {
	mut      sync.RWMutex
	data     map[string]*list.Element
	capacity int
	list     *list.List
}

func New[V any](c int) *Store[V] {

	d := make(map[string]*list.Element)

	l := list.New()

	return &Store[V]{
		data:     d,
		list:     l,
		capacity: c,
	}

}

func (s *Store[V]) Set(key string, value V) {
	s.mut.Lock()
	defer s.mut.Unlock()

	element, ok := s.data[key]
	if ok {
		// update value
		element.Value = entry[V]{key: key, value: value}
		s.list.MoveToFront(element)
	} else {
		e := &entry[V]{key: key, value: value}
		elem := s.list.PushFront(e)
		s.data[key] = elem
	}
	if len(s.data) > s.capacity {
		oldest := s.list.Back()
		e := oldest.Value.(*entry[V])
		delete(s.data, e.key)
		s.list.Remove(oldest)
	}
}

func (s *Store[V]) Get(key string) (V, bool) {
	var zero V
	s.mut.RLock()
	defer s.mut.RUnlock()

	element, ok := s.data[key]
	if ok == false {
		return zero, false
	}
	s.list.MoveToFront(element)
	e := element.Value.(*entry[V])

	return e.value, ok

}

func (s *Store[V]) Delete(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	element, ok := s.data[key]
	if ok {
		s.list.Remove(element)
		delete(s.data, key)
	}
}
