package uniqsym

import "testing"

var sunnyTests = []struct {
	name string
	str  string
	adt  ADT
}{
	{"sym only", "a", ADT{"a", nil}},
	{"sym and single-segment ns", "b.a", ADT{"a", &ADT{"b", nil}}},
	{"sym and two-segment ns", "c.b.a", ADT{"a", &ADT{"b", &ADT{"c", nil}}}},
}

func TestConvertFromStringSuccess(t *testing.T) {
	for _, test := range sunnyTests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := ConvertFromString(test.str)
			if !got.Equal(test.adt) {
				t.Errorf("got %+v, want %+v", got, test.adt)
			}
		})
	}
}

func TestConvertToStringSuccess(t *testing.T) {
	for _, test := range sunnyTests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertToString(test.adt)
			if got != test.str {
				t.Errorf("got %s, want %s", got, test.str)
			}
		})
	}
}

func TestConvertFromStringError(t *testing.T) {
	var rainyTests = []struct {
		name string
		str  string
	}{
		{"empty string", ""},
		{"sep only", "."},
		{"invalid sym", "a."},
		{"invalid ns", ".a"},
	}
	for _, test := range rainyTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ConvertFromString(test.str)
			if err == nil {
				t.Errorf("got nil error")
			}
		})
	}
}
