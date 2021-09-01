package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
)

func GenPrivateKey(bits int) []byte {
	privateKey, _ := rsa.GenerateKey(rand.Reader, bits)
	bs := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKey.Public()
	return bs
}

func ParsePubKeyWithPrivateKey(bytes []byte) []byte {
	privateKey, _ := x509.ParsePKCS1PrivateKey(bytes)
	pubKey := privateKey.PublicKey
	bs := x509.MarshalPKCS1PublicKey(&pubKey)
	return bs
}
