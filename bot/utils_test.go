package bot

import (
	"testing"
)

func TestRandStr(t *testing.T) {
	var (
		in       = 32
		expected = 32
	)
	s := RandStr(in)
	actual := len(s)
	if actual != expected {
		t.Errorf("RandStr(%d) = %d, expected %d", in, actual, expected)
	}
}
