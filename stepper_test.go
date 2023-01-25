package progressbar_test

import (
	"strconv"
	"testing"
)

func TestFloatFormat(t *testing.T) {
	fltfmt := func(f float64) string {
		return strconv.FormatFloat(f, 'f', 1, 64) + "%"
	}
	t.Logf("%v", fltfmt(0.71))
	t.Logf("%v", fltfmt(0.75))
	t.Logf("%v", fltfmt(1.01))
	t.Logf("%v", fltfmt(1.0))
}
