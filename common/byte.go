package common

import (
	"encoding/hex"
)

var Zero = int(0)

func Decode16Byte(str string) []byte {
	result, _ := hex.DecodeString(str)
	return result
}

func Encode16Byte(byteVal []byte) string {
	result := hex.EncodeToString(byteVal)
	return result
}

// func Iszero(val []byte, direction string) int {
// 	var zero [32]byte
// 	if direction == "left" {
// 		return bytes.Compare(val, zero[:])
// 	}
// 	return bytes.Compare(zero[:], val)
// }
