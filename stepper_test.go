package progressbar_test

import (
	"strconv"
	"testing"
)

func TestFloatFormat(t *testing.T) {
	fltfmt := func(f float64) string {
		return strconv.FormatFloat(f, 'f', 1, 64) + "%"
	}
	for i, cs := range []struct { //nolint:govet //can be reordered
		given  float64
		expect string
	}{
		{0.71, "0.7%"},
		{0.75, "0.8%"},
		{1.01, "1.0%"},
		{1.0, "1.0%"},
		{1.05, "1.1%"},
		{1.5, "1.5%"},
	} {
		got := fltfmt(cs.given)
		if got != cs.expect {
			t.Fatalf(`%5d. case fltfmt(%v) -> %q, but got %q`, i, cs.given, cs.expect, got)
		} else {
			t.Logf(`%5d. case fltfmt(%v) -> %q, passed.`, i, cs.given, cs.expect)
		}
	}

	// t.Logf("%v", fltfmt(0.71))
	// t.Logf("%v", fltfmt(0.75))
	// t.Logf("%v", fltfmt(1.01))
	// t.Logf("%v", fltfmt(1.0))
}
