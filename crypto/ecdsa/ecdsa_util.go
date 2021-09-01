package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
)

func GenP256PrivateKey() []byte {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	bs, _ := x509.MarshalECPrivateKey(key)
	return bs
}
func ParsePubKeyWithPrivateKey(bytes []byte) []byte {
	key, _ := x509.ParseECPrivateKey(bytes)
	pub := key.PublicKey
	bs, _ := x509.MarshalPKIXPublicKey(&pub)
	return bs
}
