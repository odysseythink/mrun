package mergo

import (
	"testing"
)

func TestIssue209(t *testing.T) {
	dst := []string{"a", "b"}
	src := []string{"c", "d"}

	if err := Merge(&dst, src, WithAppendSlice); err != nil {
		t.Error(err)
	}

	expected := []string{"a", "b", "c", "d"}
	if len(dst) != len(expected) {
		t.Errorf("arrays not equal length")
	}
	for i := range expected {
		if dst[i] != expected[i] {
			t.Errorf("array elements at %d are not equal", i)
		}
	}
}
