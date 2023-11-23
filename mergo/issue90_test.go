package mergo

import (
	"reflect"
	"testing"
)

type structWithStringMap struct {
	Data map[string]string
}

func TestIssue90(t *testing.T) {
	dst := map[string]structWithStringMap{
		"struct": {
			Data: nil,
		},
	}
	src := map[string]structWithStringMap{
		"struct": {
			Data: map[string]string{
				"foo": "bar",
			},
		},
	}
	expected := map[string]structWithStringMap{
		"struct": {
			Data: map[string]string{
				"foo": "bar",
			},
		},
	}

	err := Merge(&dst, src, WithOverride)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !reflect.DeepEqual(dst, expected) {
		t.Errorf("expected: %#v\ngot: %#v", expected, dst)
	}
}
