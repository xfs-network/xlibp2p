package p2p

import (
	"testing"
	"xlibp2p/crypto"
	"xlibp2p/discover"
)

func TestHello(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	pubKey := key.PublicKey
	id := discover.PubKey2NodeId(pubKey)
	hello := &helloRequestMsg{
		version: 1,
		id:      id,
	}
	data := hello.marshal()
	t.Logf("data: %v\n", data)
	got := new(helloRequestMsg)
	got.unmarshal(data)
	gothash := got.id
	t.Logf("gothash: %v\n", gothash)
	t.Logf("wanthash: %v\n", id)
}

func TestHello2(t *testing.T) {
	data := []byte{1, 0, 32, 0, 0, 0}
	t.Logf("data: %v\n", data)
	n := uint32(data[2]) | uint32(data[3])<<8 | uint32(data[4])<<16 | uint32(data[5])<<24
	t.Logf("n: %v\n", n)
}
