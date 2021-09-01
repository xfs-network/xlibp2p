package common

import (
	"bytes"
	"math/bits"
)

// XOR allocates a new byte slice with the computed result of XOR(a, b).
func XOR(a, b []byte) []byte {
	if len(a) != len(b) {
		return a
	}

	c := make([]byte, len(a))

	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}

	return c
}

// PrefixDiff counts the number of equal prefixed bits of a and b.
func PrefixDiff(a, b []byte, n int) int {
	buf, total := XOR(a, b), 0

	for i, b := range buf {
		if 8*i >= n {
			break
		}

		if n > 8*i && n < 8*(i+1) {
			shift := 8 - uint(n%8)
			b >>= shift
		}

		total += bits.OnesCount8(b)
	}

	return total
}

// PrefixLen returns the number of prefixed zero bits of a.
func PrefixLen(a []byte) int {
	for i, b := range a {
		if b != 0 {
			return i*8 + bits.LeadingZeros8(b)
		}
	}

	return len(a) * 8
}

func IsZero(a []byte) bool {
	z := make([]byte, len(a))
	return bytes.Compare(a, z) == Zero
}

func BytesEquals(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return bytes.Compare(a, b) == Zero
}
