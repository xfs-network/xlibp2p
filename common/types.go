package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const (
	addrLen            = 25
	hashLen            = 32
	AddrDefaultVersion = 0
)

type (
	Hash    [hashLen]byte
	Address [addrLen]byte
)

var (
	ZeroHash        = Bytes2Hash([]byte{})
	AddrCheckSumLen = 4
)

func Hex2bytes(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
		bs, _ := hex.DecodeString(s)
		return bs
	}
	return nil
}

func Bytes2Hash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func Hex2Hash(s string) Hash {
	return Bytes2Hash(Hex2bytes(s))
}

func (h *Hash) SetBytes(other []byte) {
	for i, v := range other {
		h[i] = v
	}
}

func (h *Hash) Hex() string {
	return "0x" + hex.EncodeToString(h[:])
}
func (h *Hash) Bytes() []byte {
	return h[:]
}

func IsZeroHash(h Hash) bool {
	// z := make([]byte, hashLen)
	// if bytes.Compare(h.Bytes(), z[:]) == Zero {
	// 	return true
	// }
	// return false
	var z [hashLen]byte
	return bytes.Compare(h.Bytes(), z[:]) == Zero
}

func Bytes2Address(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func B58ToAddress(enc []byte) Address {
	return Bytes2Address(B58Decode(enc))
}
func StrB58ToAddress(enc string) Address {
	return B58ToAddress([]byte(enc))
}

func Hex2Address(s string) Address {
	return Bytes2Address(Hex2bytes(s))
}

func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-addrLen:]
	}
	copy(a[addrLen-len(b):], b)
}

func (a *Address) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a *Address) Bytes() []byte {
	return a[:]
}

func (a *Address) String() string {
	return a.B58String()
}

func (a *Address) B58() []byte {
	return B58Encode(a.Bytes())
}

func (a *Address) Version() uint8 {
	return a[0]
}
func (a *Address) PubKeyHash() []byte {
	return a[1 : addrLen-AddrCheckSumLen]
}
func (a *Address) Payload() []byte {
	return a[:addrLen-AddrCheckSumLen]
}
func (a *Address) Checksum() []byte {
	return a[1+(addrLen-AddrCheckSumLen)-1:]
}

func (a *Address) B58String() string {
	return string(a.B58())
}

func (a *Address) Equals(b Address) bool {
	return bytes.Compare(a.Bytes(), b.Bytes()) == Zero
}

func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", h.Hex())), nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	if data == nil || len(data) < hashLen {
		h.SetBytes([]byte{0})
		return nil
	}
	hash := Hex2Hash(string(data[1 : len(data)-1]))
	h.SetBytes(hash.Bytes())
	return nil
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", a.B58())), nil
}

func (a *Address) UnmarshalJSON(data []byte) error {
	if data == nil || len(data) < 28 {
		a.SetBytes([]byte{0})
		return nil
	}
	b58a := B58ToAddress(data[1 : len(data)-1])
	a.SetBytes(b58a.Bytes())
	return nil
}
