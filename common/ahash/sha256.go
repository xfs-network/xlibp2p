package ahash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	var bs = make([]byte, len(hash))
	copy(bs, hash[:])
	return bs
}

func SHA256HEX(data []byte) string {
	return hex.EncodeToString(SHA256(data))
}
