package containers

import "testing"

func TestSetNew(t *testing.T) {
	var sm SyncMap[string, string]
	sm.Store("test", "")
	sm.Store("Hello", "world")

	if val, ok := sm.Load("test"); !ok || val != "" {
		t.Errorf("Got %v expected empty", val)
	}

	if val, ok := sm.Load("Hello"); !ok || val != "world" {
		t.Errorf("Got %v expected empty", val)
	}

	sm.Store("test", "123")
	if val, ok := sm.Load("test"); !ok || val != "123" {
		t.Errorf("Got %v expected empty", val)
	}
}
