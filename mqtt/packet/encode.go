package packet

import "encoding/binary"

func encodeString(str string) []byte {
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(str)))
	return append(length, []byte(str)...)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func int2bool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}
