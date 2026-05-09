package store

import "testing"

// A new store starts empty with the given capacity.
func TestNewStore(t *testing.T) {
	s := New[string](10)
	if s.capacity != 10 {
		t.Fatalf("expected capacity 10, got %d", s.capacity)
	}
	if len(s.data) != 0 {
		t.Fatalf("expected empty data, got %d entries", len(s.data))
	}
}

// Setting and getting a string value returns what was stored.
func TestSetAndGetString(t *testing.T) {
	s := New[string](10)
	s.Set("name", "nexo")
	val, ok := s.Get("name")
	if !ok {
		t.Fatal("expected ok=true for existing key")
	}
	if val != "nexo" {
		t.Fatalf("expected 'nexo', got %q", val)
	}
}

// The store works with any type — int values are stored and retrieved.
func TestSetAndGetInt(t *testing.T) {
	s := New[int](10)
	s.Set("count", 42)
	val, ok := s.Get("count")
	if !ok {
		t.Fatal("expected ok=true for existing key")
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

// Getting a key that doesn't exist returns the zero value and false.
func TestGetMissingKeyReturnsZeroValue(t *testing.T) {
	s := New[string](10)
	val, ok := s.Get("nothing")
	if ok {
		t.Fatal("expected ok=false for missing key")
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
}

// After deleting a key, Get no longer finds it.
func TestDeleteRemovesKey(t *testing.T) {
	s := New[string](10)
	s.Set("name", "nexo")
	s.Delete("name")
	_, ok := s.Get("name")
	if ok {
		t.Fatal("expected ok=false after delete")
	}
}

// Deleting a key that doesn't exist should not panic.
func TestDeleteNonExistentKeyDoesNotPanic(t *testing.T) {
	s := New[string](10)
	s.Delete("nothing")
}

// Setting the same key again updates its value and moves it to the front.
func TestSetUpdateExistingKey(t *testing.T) {
	s := New[string](10)
	s.Set("name", "first")
	s.Set("name", "second")
	val, ok := s.Get("name")
	if !ok {
		t.Fatal("expected ok=true after update")
	}
	if val != "second" {
		t.Fatalf("expected 'second', got %q", val)
	}
}

// Adding items beyond capacity evicts the oldest (least recently used) item.
func TestEvictionRemovesOldest(t *testing.T) {
	s := New[string](2)
	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")

	_, ok := s.Get("a")
	if ok {
		t.Fatal("expected 'a' to be evicted (oldest)")
	}
}

// Recently accessed items stay — eviction removes the unaccessed one.
func TestGetRefreshesLRUOrder(t *testing.T) {
	s := New[string](2)
	s.Set("a", "1")
	s.Set("b", "2")
	s.Get("a")
	s.Set("c", "3")

	_, ok := s.Get("a")
	if !ok {
		t.Fatal("expected 'a' to still exist (was refreshed by Get)")
	}
	_, ok2 := s.Get("b")
	if ok2 {
		t.Fatal("expected 'b' to be evicted (was not refreshed)")
	}
	_, ok3 := s.Get("c")
	if !ok3 {
		t.Fatal("expected 'c' to exist (just added)")
	}
}
