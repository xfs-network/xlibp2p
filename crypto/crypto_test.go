package crypto

import (
	"errors"
	"fmt"
	"github.com/xfs-network/xlibp2p/common"
	"testing"
)

func TestPubKey2Addr(t *testing.T) {
	prv, err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := DefaultPubKey2Addr(prv.PublicKey)
	t.Logf("addr: %s\n", addr.String())
	keyEnc, err := PrivateKeyEncodeB64String(prv)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("privateKey: %s\n", keyEnc)
	got := common.StrB58ToAddress(addr.String())

	if !addr.Equals(got) {
		t.Fatal(fmt.Errorf("not equals"))
	}
	if !VerifyAddress(addr) {
		t.Fatal(fmt.Errorf("not verify"))
	}
}

func TestPubKeyEncode(t *testing.T) {
	key,err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	enc := PubKeyEncode(key.PublicKey)
	t.Logf("len: %d, %x", len(enc), enc)
}

func TestVerifyAddress(t *testing.T) {
	key,err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	p := key.PublicKey
	if p.Curve == nil || p.X == nil || p.Y == nil {
		t.Fatal(errors.New("nil err"))
	}
	xbs := p.X.Bytes()
	ybs := p.Y.Bytes()
	buf := make([]byte, len(xbs) + len(ybs))
	copy(buf,append(xbs,ybs...))
	t.Logf("x len: %d, %x", len(xbs), xbs)
	t.Logf("y len: %d, %x", len(ybs), ybs)
	t.Logf("buf len: %d, %x", len(buf), buf)
}
