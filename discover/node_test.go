package discover

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strconv"
	"testing"
	"xlibp2p/crypto"
)


func TestPubKey2NodeId(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	wantBuf := crypto.PubKeyEncode(key.PublicKey)
	gotId := PubKey2NodeId(key.PublicKey)
	if !bytes.Equal(wantBuf,gotId[:]){
		t.Fatalf("got id: %s, want: %x", gotId, wantBuf)
	}
}

func TestHex2NodeId(t *testing.T) {
	ref := NodeId{
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,36,22,98,66,128,214,1}
	hex := "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002416624280d601"
	nid, err := Hex2NodeId(hex)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(ref[:], nid[:]) {
		t.Fatalf("got id: %s, want: %s", nid, ref)
	}
}

func TestParseNode(t *testing.T) {
	nid := NodeId{
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,36,22,98,66,128,214,1}
	nidHex := "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002416624280d601"
	wantNode := newNode(net.IP{127,0,0,1},52150,22334, nid)
	rawUrl := fmt.Sprintf("xfsnode://%s@127.0.0.1:52150?discport=22334", nidHex)
	t.Logf("parse raw url: %s", rawUrl)
	got,err := ParseNode(rawUrl)

	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wantNode.IP, got.IP) {
		t.Fatalf("got ip: %s, want: %s", got.IP, wantNode.IP)
	}
	if wantNode.TCP != got.TCP {
		t.Fatalf("got tcp port: %d, want: %d", got.TCP, wantNode.TCP)
	}
	if wantNode.UDP != got.UDP {
		t.Fatalf("got udp port: %d, want: %d", got.UDP, wantNode.UDP)
	}
	if !bytes.Equal(wantNode.ID[:],got.ID[:]) {
		t.Fatalf("got id: %s, want: %s", got.ID, wantNode.ID)
	}
	if !bytes.Equal(wantNode.Hash[:], got.Hash[:]){
		t.Fatalf("got hash: %s, want: %s", got.ID, wantNode.ID)
	}
}


func TestNode_logdist(t *testing.T) {
	logdistBig:= func(a,b []byte) int {
		aBig := new(big.Int).SetBytes(a)
		bBig := new(big.Int).SetBytes(b)
		return new(big.Int).Xor(aBig,bBig).BitLen()
	}
	var (
		a = [2]byte{1,2}
		b = [2]byte{3,4}
	)
	if n := logdist(a[:],a[:]); n != 0{
		t.Fatalf("got self logdist: %d, want: 0", n)
	}
	want := logdistBig(a[:],b[:])
	got := logdist(a[:],b[:])
	t.Logf("logdist: %d", got)
	if want != got {
		t.Fatalf("got logdist: %d, want: %d", got, want)
	}
}

func TestNode_distcmp(t *testing.T) {
	distcmpBig := func(target, a, b []byte) int {
		tbig := new(big.Int).SetBytes(target)
		abig := new(big.Int).SetBytes(a)
		bbig := new(big.Int).SetBytes(b)
		return new(big.Int).Xor(tbig, abig).Cmp(new(big.Int).Xor(tbig, bbig))
	}
	var (
		a = [2]byte{1,2}
		b = [2]byte{3,4}
		c = [2]byte{3,4}
	)
	if n := distcmp(a[:],a[:],a[:]); n != 0{
		t.Fatalf("got self logdist: %d, want: 0", n)
	}
	want := distcmpBig(a[:],b[:],c[:])
	got := distcmp(a[:],b[:],c[:])
	if want != got {
		t.Fatalf("got distcmp: %d, want: %d", got, want)
	}
}

func TestParseNode2(t *testing.T) {
	str := "xfsnode://127.0.0.1:9091/?id=8835c3a73333e8bf26eb28b3fd958f68ec32b0cd8c7e1fcdc090b2f3cdabd39fc7a5c5e23994cc74d60db5ab41163e966ccf09883fb112fc4f476c06e19035e9"
	parseNode := func(rawurl string) (*Node, error) {
		var (
			id NodeId
			ip net.IP
			tcpPort, udpPort uint64
		)
		u, err := url.Parse(rawurl)
		if err != nil {
			t.Fatal(err)
		}
		if u.Scheme != "xfsnode" {
			return nil, errors.New("invalid URL scheme, want \"xfsnode\"")
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return nil, fmt.Errorf("invalid host: %v", err)
		}
		if ip = net.ParseIP(host); ip == nil {
			return nil, errors.New("invalid ip host")
		}
		if ipv4 := ip.To4(); ipv4 != nil {
			ip = ipv4
		}
		if tcpPort, err = strconv.ParseUint(port, 10, 16); err != nil {
			return nil, errors.New("invalid port")
		}
		udpPort = tcpPort
		q := u.Query()
		nId := q.Get("id")
		if nId == "" {
			return nil, errors.New("does not contain node ID")
		}
		if id, err = Hex2NodeId(nId); err != nil {
			return nil, fmt.Errorf("invalid node ID (%v)", err)
		}
		return newNode(ip, uint16(tcpPort), uint16(udpPort), id), nil
	}
	n, err := parseNode(str)
	if err != nil {
		t.Fatal(err)
	}
	_=n
}