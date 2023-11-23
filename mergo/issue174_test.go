package mergo

import (
	"testing"
)

type structWithBlankField struct {
	_ struct{}
	A struct{}
}

func TestIssue174(t *testing.T) {
	dst := structWithBlankField{}
	src := structWithBlankField{}

	if err := Merge(&dst, src, WithOverride); err != nil {
		t.Error(err)
	}
}
