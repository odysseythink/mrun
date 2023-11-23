package mergo

import (
	"reflect"
	"testing"
)

func TestIssue220(t *testing.T) {
	dst := []interface{}{
		map[string]int{
			"a": 1,
		},
	}
	src := []interface{}{
		"nil",
	}
	expected := []interface{}{
		map[string]int{
			"a": 1,
		},
	}

	err := Merge(&dst, src, WithSliceDeepCopy)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !reflect.DeepEqual(dst, expected) {
		t.Errorf("expected: %#v\ngot: %#v", expected, dst)
	}
}
