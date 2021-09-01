package ahash

import "golang.org/x/crypto/ripemd160"

func Ripemd160(data []byte) []byte {
	hashc := ripemd160.New()
	hashc.Write(data)
	return hashc.Sum(nil)
}
