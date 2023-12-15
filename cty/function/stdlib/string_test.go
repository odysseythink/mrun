package stdlib

import (
	"fmt"
	"testing"

	"mlib.com/mrun/cty"
)

func TestUpper(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal("HELLO"),
			cty.StringVal("HELLO"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("жж"),
			cty.StringVal("ЖЖ"),
		},
		{
			cty.StringVal("noël"),
			cty.StringVal("NOËL"),
		},
		{
			// Go's case conversions don't handle this ligature, which is
			// unfortunate but is now a compatibility constraint since it
			// would be potentially-breaking to behave differently here in
			// future.
			cty.StringVal("baﬄe"),
			cty.StringVal("BAﬄE"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😸😾"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.StringVal("hello").Mark(1),
			cty.StringVal("HELLO").Mark(1),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Upper(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("HELLO"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("ЖЖ"),
			cty.StringVal("жж"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Lower(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("olleh"),
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
		},
		{
			cty.StringVal("1"),
			cty.StringVal("1"),
		},
		{
			cty.StringVal("Живой Журнал"),
			cty.StringVal("ланруЖ йовиЖ"),
		},
		{
			// note that the dieresis here is intentionally a combining
			// ligature.
			cty.StringVal("noël"),
			cty.StringVal("lëon"),
		},
		{
			// The Es in this string has three combining acute accents.
			// This tests something that NFC-normalization cannot collapse
			// into a single precombined codepoint, since otherwise we might
			// be cheating and relying on the single-codepoint forms.
			cty.StringVal("wé́́é́́é́́!"),
			cty.StringVal("!é́́é́́é́́w"),
		},
		{
			// Go's normalization forms don't handle this ligature, so we
			// will produce the wrong result but this is now a compatibility
			// constraint and so we'll test it.
			cty.StringVal("baﬄe"),
			cty.StringVal("eﬄab"),
		},
		{
			cty.StringVal("😸😾"),
			cty.StringVal("😾😸"),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Reverse(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestStrlen(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(5),
		},
		{
			cty.StringVal(""),
			cty.NumberIntVal(0),
		},
		{
			cty.StringVal("1"),
			cty.NumberIntVal(1),
		},
		{
			cty.StringVal("Живой Журнал"),
			cty.NumberIntVal(12),
		},
		{
			// note that the dieresis here is intentionally a combining
			// ligature.
			cty.StringVal("noël"),
			cty.NumberIntVal(4),
		},
		{
			// The Es in this string has three combining acute accents.
			// This tests something that NFC-normalization cannot collapse
			// into a single precombined codepoint, since otherwise we might
			// be cheating and relying on the single-codepoint forms.
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(5),
		},
		{
			// Go's normalization forms don't handle this ligature, so we
			// will produce the wrong result but this is now a compatibility
			// constraint and so we'll test it.
			cty.StringVal("baﬄe"),
			cty.NumberIntVal(4),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(2),
		},
		{
			cty.UnknownVal(cty.String),
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.Zero, true).NewValue(),
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("wé́́é́́é́́-").NewValue(),
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.NumberIntVal(5), true).NewValue(),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number).Refine().NotNull().NumberRangeLowerBound(cty.Zero, true).NewValue(),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Strlen(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSubstr(t *testing.T) {
	tests := []struct {
		Input  cty.Value
		Offset cty.Value
		Length cty.Value
		Want   cty.Value
	}{
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(2),
			cty.StringVal("he"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(3),
			cty.StringVal("ell"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(-1),
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(-10), // not documented, but <0 is the same as -1
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(10),
			cty.StringVal("ello"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(-3),
			cty.NumberIntVal(-1),
			cty.StringVal("llo"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(-3),
			cty.NumberIntVal(2),
			cty.StringVal("ll"),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(10),
			cty.NumberIntVal(10),
			cty.StringVal(""),
		},
		{
			cty.StringVal("hello"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(0),
			cty.StringVal(""),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(3),
			cty.StringVal("noë"),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(3),
			cty.NumberIntVal(-1),
			cty.StringVal("l"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(2),
			cty.NumberIntVal(2),
			cty.StringVal("é́́é́́"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(3),
			cty.NumberIntVal(2),
			cty.StringVal("é́́!"),
		},
		{
			cty.StringVal("wé́́é́́é́́!"),
			cty.NumberIntVal(-2),
			cty.NumberIntVal(-1),
			cty.StringVal("é́́!"),
		},
		{
			cty.StringVal("noël"),
			cty.NumberIntVal(-2),
			cty.NumberIntVal(-1),
			cty.StringVal("ël"),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(0),
			cty.NumberIntVal(1),
			cty.StringVal("😸"),
		},
		{
			cty.StringVal("😸😾"),
			cty.NumberIntVal(1),
			cty.NumberIntVal(1),
			cty.StringVal("😾"),
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			got, err := Substr(test.Input, test.Offset, test.Length)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := map[string]struct {
		Separator cty.Value
		Lists     []cty.Value
		Want      cty.Value
	}{
		"single two-element list": {
			cty.StringVal("-"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
			},
			cty.StringVal("hello-world"),
		},
		"multiple single-element lists": {
			cty.StringVal("-"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("chicken")}),
				cty.ListVal([]cty.Value{cty.StringVal("egg")}),
			},
			cty.StringVal("chicken-egg"),
		},
		"single single-element list": {
			cty.StringVal("-"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("chicken")}),
			},
			cty.StringVal("chicken"),
		},
		"blank separator": {
			cty.StringVal(""),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("horse"), cty.StringVal("face")}),
			},
			cty.StringVal("horseface"),
		},
		"marked list": {
			cty.StringVal("-"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}).Mark("sensitive"),
			},
			cty.StringVal("hello-world").Mark("sensitive"),
		},
		"marked separator": {
			cty.StringVal("-").Mark("sensitive"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}),
			},
			cty.StringVal("hello-world").Mark("sensitive"),
		},
		"list with some marked elements": {
			cty.StringVal("-"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello").Mark("sensitive"), cty.StringVal("world")}),
			},
			cty.StringVal("hello-world").Mark("sensitive"),
		},
		"multiple marks": {
			cty.StringVal("-").Mark("a"),
			[]cty.Value{
				cty.ListVal([]cty.Value{cty.StringVal("hello").Mark("b"), cty.StringVal("world").Mark("c")}),
			},
			cty.StringVal("hello-world").WithMarks(cty.NewValueMarks("a", "b", "c")),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Join(test.Separator, test.Lists...)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		Input   cty.Value
		Want    cty.Value
		WantErr string
	}{
		{
			cty.ListValEmpty(cty.String),
			cty.ListValEmpty(cty.String),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("a")}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("b"), cty.StringVal("a")}),
			cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			``,
		},
		{
			cty.ListVal([]cty.Value{cty.StringVal("b"), cty.StringVal("a"), cty.StringVal("c")}),
			cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("c")}),
			``,
		},
		{
			cty.UnknownVal(cty.List(cty.String)),
			cty.UnknownVal(cty.List(cty.String)).RefineNotNull(),
			``,
		},
		{
			// If the list contains any unknown values then we can still
			// preserve the length of the list by generating a known list
			// with unknown elements, because sort can never change the length.
			cty.ListVal([]cty.Value{cty.StringVal("b"), cty.UnknownVal(cty.String)}),
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.String), cty.UnknownVal(cty.String)}),
			``,
		},
		{
			// For a completely unknown list we can still preserve any
			// refinements it had for its length, because sorting can never
			// change the length.
			cty.UnknownVal(cty.List(cty.String)).Refine().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(2).
				NewValue(),
			cty.UnknownVal(cty.List(cty.String)).Refine().
				NotNull().
				CollectionLengthLowerBound(1).
				CollectionLengthUpperBound(2).
				NewValue(),
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Sort(%#v)", test.Input), func(t *testing.T) {
			got, err := Sort(test.Input)

			if test.WantErr != "" {
				errStr := fmt.Sprintf("%s", err)
				if errStr != test.WantErr {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", errStr, test.WantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)
			}
		})
	}
}
