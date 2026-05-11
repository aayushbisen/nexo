package store

import (
	"container/list"
	"sync"
	"time"
)

type entry[V any] struct {
	key       string
	value     V
	expiresAt time.Time
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

func (s *Store[V]) Set(key string, value V, ttl ...time.Duration) {
	s.mut.Lock()
	defer s.mut.Unlock()

	var expiresAt time.Time
	if len(ttl) > 0 && ttl[0] > 0 {
		expiresAt = time.Now().Add(ttl[0])
	}

	element, ok := s.data[key]
	if ok {
		// update value
		element.Value = &entry[V]{key: key, value: value, expiresAt: expiresAt}
		s.list.MoveToFront(element)
	} else {
		e := &entry[V]{key: key, value: value, expiresAt: expiresAt}
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
	s.mut.Lock()
	defer s.mut.Unlock()

	element, ok := s.data[key]
	if !ok {
		return zero, false
	}

	e := element.Value.(*entry[V])
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		s.list.Remove(element)
		delete(s.data, key)
		return zero, false
	}

	s.list.MoveToFront(element)
	return e.value, true
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

func (s *Store[V]) Expire(key string, ttl time.Duration) bool {
	s.mut.Lock()
	defer s.mut.Unlock()

	element, ok := s.data[key]
	if !ok {
		return false
	}

	e := element.Value.(*entry[V])
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		s.list.Remove(element)
		delete(s.data, key)
		return false
	}

	e.expiresAt = time.Now().Add(ttl)
	element.Value = e
	s.list.MoveToFront(element)
	return true
}
