package bot

import (
	"testing"
)

func TestGetJsonStrValue(t *testing.T) {
	var (
		in       = []byte("{\"hello\": {\"go\": \"world\"}}")
		expected = "world"
	)
	actual := GetJsonStrVal(in, "hello", "go")
	if actual != expected {
		t.Errorf("GetJsonStrValue(%s) = %s, expected %s", in, actual, expected)
	}
}

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
