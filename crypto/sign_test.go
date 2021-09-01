package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"xlibp2p/common/ahash"
)

func TestSign(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	hash := ahash.SHA256([]byte("i"))
	sig, err := ECDSASign(hash, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("signature: %x\n", sig)
	if !VerifySignature(hash, sig) {
		t.Fatal("not verify")
	}
}
