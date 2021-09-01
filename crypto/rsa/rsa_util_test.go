package rsa

import (
	"encoding/base64"
	"testing"
)

func TestGenRSAPrivateKey(t *testing.T) {
	bs := GenPrivateKey(2048)
	pub := ParsePubKeyWithPrivateKey(bs)
	//bss := Bytes2Hex(bs)
	bsa := base64.StdEncoding.EncodeToString(pub)
	t.Logf("k: %s\n", bsa)
}
