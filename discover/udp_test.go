package discover

import (
	"bytes"
	"net"
	"testing"
	"time"
	"xlibp2p/crypto"
)


func TestListenUDP(t *testing.T) {

	//test.udp.
}

func Test_encodePacket(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	selfAddr := net.IP{127,0,0,1}
	selfEnd := rpcEndpoint{
		IP: selfAddr,
		UDP: 9001,
		TCP: 9001,
	}
	targetAddr := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9002,
	}
	pingPacketObj := ping{
		Version: Version,
		From: selfEnd,
		To: makeEndpoint(targetAddr, 0),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}
	raw, err := encodePacket(key, pingPacket, pingPacketObj)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode raw: %x", raw)
	gotType := raw[0]
	if gotType != pingPacket{
		t.Fatalf("got packet type: %d, want: %x", gotType, pingPacket)
	}
	wantId := PubKey2NodeId(key.PublicKey)
	gotId := raw[1:nodeIdLen+1]
	t.Logf("nodeId raw: %x", gotId)
	if !bytes.Equal(wantId[:], gotId) {
		t.Fatalf("got node id: %x, want: %x", gotId, wantId[:])
	}
	data := raw[1 + len(gotId):]
	t.Logf("data raw: %x", data)
	t.Logf("data string: %s", string(data))
}
func assertRpcEndpoint(t *testing.T, tag string, got *rpcEndpoint, want *rpcEndpoint)  {
	if !bytes.Equal(got.IP.To4()[:], want.IP.To4()[:]) {
		t.Fatalf("got %s ip addr: %s, want: %s", tag, got.IP, want.IP)
	}
	if got.TCP != want.TCP {
		t.Fatalf("got %s tcp port: %d, want: %d",tag, got.TCP, want.TCP)
	}
	if got.UDP != want.UDP {
		t.Fatalf("got %s udp port: %d, want: %d",tag, got.UDP, want.UDP)
	}
}
func Test_decodePacket(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	selfAddr := net.IP{127,0,0,1}.To4()
	selfEnd := rpcEndpoint{
		IP: selfAddr,
		UDP: 9001,
		TCP: 9001,
	}
	targetAddr,err := net.ResolveUDPAddr("","129.0.0.1:9091")
	if err != nil {
		t.Fatal(err)
	}
	pingPacketObj := ping{
		Version: Version,
		From: selfEnd,
		To: makeEndpoint(targetAddr, 0),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	}
	raw, err := encodePacket(key, pingPacket, pingPacketObj)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encode raw: %x", raw)
	selfId := PubKey2NodeId(key.PublicKey)
	mBuffer := bytes.NewBuffer(raw)
	pack,nId,err := decodePacket(mBuffer)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(nId[:], selfId[:]) {
		t.Fatalf("got node id: %x, want: %x", nId[:], selfId[:])
	}
	gotPack := pack.(*ping)
	if gotPack.Version != pingPacketObj.Version {
		t.Fatalf("got packet version: %d, want: %d", gotPack.Version, pingPacketObj.Version)
	}
	if gotPack.Expiration != pingPacketObj.Expiration {
		t.Fatalf("got packet expiration: %d, want: %d", gotPack.Expiration, pingPacketObj.Expiration)
	}
	if gotPack.Expiration != pingPacketObj.Expiration {
		t.Fatalf("got packet expiration: %d, want: %d", gotPack.Expiration, pingPacketObj.Expiration)
	}
	assertRpcEndpoint(t,"from", &gotPack.From, &pingPacketObj.From)
	assertRpcEndpoint(t,"to", &gotPack.To, &pingPacketObj.To)
}