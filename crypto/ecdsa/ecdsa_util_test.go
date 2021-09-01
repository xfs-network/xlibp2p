package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func TestGenP256PrivateKey(t *testing.T) {
	key := GenP256PrivateKey()
	pub := ParsePubKeyWithPrivateKey(key)
	bas := base64.StdEncoding.EncodeToString(pub)
	t.Logf("k: %s\n", bas)
}

func TestSign(t *testing.T) {
	key := GenP256PrivateKey()
	basKey := base64.StdEncoding.EncodeToString(key)
	t.Logf("key: %s\n", basKey)

	pub := ParsePubKeyWithPrivateKey(key)
	basPub := base64.StdEncoding.EncodeToString(pub)
	t.Logf("pub: %s\n", basPub)

	privateKey, err := x509.ParseECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("def")
	signature, err := privateKey.Sign(rand.Reader, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	basSignature := base64.StdEncoding.EncodeToString(signature)
	t.Logf("signature: %s\n", basSignature)

	pkGen, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	pk := pkGen.(*ecdsa.PublicKey)
	bl := ecdsa.VerifyASN1(pk, data, signature)
	if !bl {
		t.Fatal("VerifyASN1 err")
	}
}
