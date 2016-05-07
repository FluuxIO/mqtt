package mqtt

import (
	"bytes"
	"testing"
)

func TestDecodeRLength(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0)
	l, _ := readRemainingLength(buf)
	if l != 0 {
		t.Error("incorrect remaining length")
	}

	buf = bytes.NewBuffer(nil)
	buf.WriteByte(64)
	l, _ = readRemainingLength(buf)
	if l != 64 {
		t.Error("incorrect remaining length")
	}

	buf = bytes.NewBuffer(nil)
	buf.Write([]byte{193, 2})
	l, _ = readRemainingLength(buf)
	if l != 321 {
		t.Error("incorrect remaining length")
	}
}
