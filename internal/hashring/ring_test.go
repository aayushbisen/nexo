package hashring

import "testing"

func TestRingCount(t *testing.T) {
	r1 := New()
	r1.AddNode("locahost:9008", 5)
	r1.AddNode("locahost:9018", 5)
	want := 10
	got := len(r1.nodes)
	if got != want {
		t.Fatalf("len nodes got %d, want %d", got, want)
	}
}
