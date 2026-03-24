package cache

import "testing"

func TestLRUEvicts(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")
	c.Put("c", 3)
	if _, ok := c.Get("b"); ok {
		t.Fatalf("expected b to be evicted")
	}
}
