package progressbar

import (
	"testing"
)

func TestHumanizeBytes(t *testing.T) {
	var a, b string
	a, b = humanizeBytes(1.0)
	t.Log(a + b)
	a, b = humanizeBytes(3.739)
	t.Log(a + b)
	a, b = humanizeBytes(21)
	t.Log(a + b)
	a, b = humanizeBytes(1300.3)
	t.Log(a + b)
}

func TestConcatTwoStrings(t *testing.T) {
	t.Log("a" + "b")
}
