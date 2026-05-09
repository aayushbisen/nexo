package hashring

import "testing"

// A freshly created ring has no nodes and no mappings.
func TestNewRingIsEmpty(t *testing.T) {
	r := New()
	if len(r.nodes) != 0 {
		t.Fatalf("expected empty nodes, got %d", len(r.nodes))
	}
	if len(r.nodeMap) != 0 {
		t.Fatalf("expected empty nodeMap, got %d", len(r.nodeMap))
	}
}

// Adding one node with 5 replicas creates 5 entries in the ring.
func TestAddNodeCreatesVirtualNodes(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	if len(r.nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(r.nodes))
	}
	if len(r.nodeMap) != 5 {
		t.Fatalf("expected 5 mappings, got %d", len(r.nodeMap))
	}
}

// All 5 virtual nodes for "localhost:9091" point back to the same physical address.
func TestAddNodeAllVirtualNodesMapToSameAddress(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	for _, addr := range r.nodeMap {
		if addr != "localhost:9091" {
			t.Fatalf("expected all virtual nodes to map to localhost:9091, got %s", addr)
		}
	}
}

// Two physical nodes with 5 replicas each create 10 total virtual nodes.
func TestAddNodeMultiplePhysicalNodes(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	r.AddNode("localhost:9092", 5)
	if len(r.nodes) != 10 {
		t.Fatalf("expected 10 nodes, got %d", len(r.nodes))
	}
}

// The "$0", "$1" suffix pattern ensures every virtual node has a unique hash.
func TestAddNodeVirtualHashesAreUnique(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	seen := make(map[uint32]bool)
	for _, h := range r.nodes {
		if seen[h] {
			t.Fatalf("duplicate hash %d in ring", h)
		}
		seen[h] = true
	}
}

// Zero replicas is invalid — AddNode should panic to catch the bug early.
func TestAddNodePanicsOnZeroReplicas(t *testing.T) {
	r := New()
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("expected panic for replicas=0, got none")
		}
	}()
	r.AddNode("localhost:9091", 0)
}

// Negative replicas would cause a runtime panic — guard catches it early.
func TestAddNodePanicsOnNegativeReplicas(t *testing.T) {
	r := New()
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("expected panic for replicas=-1, got none")
		}
	}()
	r.AddNode("localhost:9091", -1)
}

// Any key should map to one of the known physical addresses in the ring.
func TestGetNodeReturnsPhysicalAddress(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	r.AddNode("localhost:9092", 5)

	addr := r.GetNode("some-key")
	if addr != "localhost:9091" && addr != "localhost:9092" {
		t.Fatalf("expected address to be one of the known nodes, got %q", addr)
	}
}

// The same key always maps to the same node — consistent hashing guarantee.
func TestGetNodeDeterministic(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	r.AddNode("localhost:9092", 5)

	a1 := r.GetNode("hello")
	a2 := r.GetNode("hello")
	if a1 != a2 {
		t.Fatalf("expected same key to map to same node every time, got %q then %q", a1, a2)
	}
}

// A key that hashes past the last node wraps around to the first node.
func TestGetNodeWrapAround(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)

	addr := r.GetNode("zzzzzzzzzzzzzzzzzzzz")
	if addr != "localhost:9091" {
		t.Fatalf("expected wrap-around to return the only node, got %q", addr)
	}
}

// With 2 physical nodes, 1000 keys should spread across both — not all to one.
func TestGetNodeDistribution(t *testing.T) {
	r := New()
	r.AddNode("localhost:9091", 5)
	r.AddNode("localhost:9092", 5)

	counts := map[string]int{"localhost:9091": 0, "localhost:9092": 0}
	for i := 0; i < 1000; i++ {
		key := "key-" + string(rune(i))
		addr := r.GetNode(key)
		counts[addr]++
	}
	if counts["localhost:9091"] == 0 {
		t.Fatal("node 9091 got 0 keys — distribution is broken")
	}
	if counts["localhost:9092"] == 0 {
		t.Fatal("node 9092 got 0 keys — distribution is broken")
	}
	if counts["localhost:9091"] == 1000 || counts["localhost:9092"] == 1000 {
		t.Fatal("one node got all keys — ring is not distributing")
	}
}
