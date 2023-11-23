package mergo

import (
	"fmt"
	"testing"
)

func TestIssue143(t *testing.T) {
	testCases := []struct {
		expected func(map[string]interface{}) error
		options  []func(*Config)
	}{
		{
			options: []func(*Config){WithOverride},
			expected: func(m map[string]interface{}) error {
				properties := m["properties"].(map[string]interface{})
				if properties["field1"] != "wrong" {
					return fmt.Errorf("expected %q, got %v", "wrong", properties["field1"])
				}
				return nil
			},
		},
		{
			options: []func(*Config){},
			expected: func(m map[string]interface{}) error {
				properties := m["properties"].(map[string]interface{})
				if properties["field1"] == "wrong" {
					return fmt.Errorf("expected a map, got %v", "wrong")
				}
				return nil
			},
		},
	}
	for _, tC := range testCases {
		base := map[string]interface{}{
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type": "text",
				},
			},
		}

		err := Map(
			&base,
			map[string]interface{}{
				"properties": map[string]interface{}{
					"field1": "wrong",
				},
			},
			tC.options...,
		)
		if err != nil {
			t.Error(err)
		}
		if err := tC.expected(base); err != nil {
			t.Error(err)
		}
	}
}
