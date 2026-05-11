package jcs

import "testing"

func TestMarshal_KeysSorted(t *testing.T) {
	in := map[string]any{"b": 2, "a": 1, "c": map[string]any{"y": "Y", "x": "X"}}
	out, err := Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":1,"b":2,"c":{"x":"X","y":"Y"}}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestSha256OfJCS_Stable(t *testing.T) {
	in1 := map[string]any{"a": 1, "b": 2}
	in2 := map[string]any{"b": 2, "a": 1}
	h1, _ := Sha256OfJCS(in1)
	h2, _ := Sha256OfJCS(in2)
	if h1 != h2 {
		t.Errorf("expected stable digest, got %s vs %s", h1, h2)
	}
}
