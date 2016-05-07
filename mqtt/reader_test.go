package mqtt

import (
	"bytes"
	"testing"
)

func TestReadRemainingLength(t *testing.T) {
	bufferCheck([]byte{0}, 0, t)
	bufferCheck([]byte{64}, 64, t)
	bufferCheck([]byte{193, 2}, 321, t)
}

func bufferCheck(input []byte, expected int, t *testing.T) {
	buf := bytes.NewBuffer(input)
	l, _ := readRemainingLength(buf)
	if l != expected {
		t.Errorf("incorrect remaining length (%d) = %d", l, expected)
	}
}
