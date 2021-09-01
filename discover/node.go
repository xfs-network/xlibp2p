package discover

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"xlibp2p/common"
	"xlibp2p/crypto"
)


const nodeIdLen int = 64

type NodeId [nodeIdLen]byte

func (id NodeId) String() string {
	return fmt.Sprintf("%x", id[:])
}

func Hex2NodeId(s string) (NodeId, error) {
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	var id NodeId
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	} else if len(b) != len(id) {
		return id, fmt.Errorf("wrong length, need %d hex bytes", len(id))
	}
	copy(id[:], b)
	return id, nil
}
func MustHex2NodeId(in string) NodeId {
	id, err := Hex2NodeId(in)
	if err != nil {
		panic(err)
	}
	return id
}
func PubKey2NodeId(pub ecdsa.PublicKey) NodeId {
	var id NodeId
	pHashBytes := crypto.PubKeyEncode(pub)
	copy(id[:], pHashBytes)
	return id
}
type Node struct {
	IP net.IP
	TCP,UDP uint16
	ID NodeId
	Hash common.Hash
}

func NewNode(ip net.IP, tcpPort, udpPort uint16, id NodeId) *Node {
	return newNode(ip,tcpPort,udpPort,id)
}
func newNode(ip net.IP, tcpPort, udpPort uint16, id NodeId) *Node {
	n :=  &Node{
		IP: ip,
		TCP: tcpPort,
		UDP: udpPort,
		ID:  id,
	}
	n.Hash = crypto.ByteHash256(id[:])
	return n
}
func (n *Node) TcpAddr() *net.TCPAddr {
	return &net.TCPAddr{IP: n.IP, Port: int(n.TCP)}
}

func (n *Node) UdpAddr() *net.UDPAddr {
	return n.addr()
}

func (n *Node) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP, Port: int(n.UDP)}
}

func (n *Node) String() string {
	addr := net.TCPAddr{IP: n.IP, Port: int(n.TCP)}
	u := url.URL{
		Scheme: "xfsnode",
		Host:   addr.String(),
	}
	u.RawQuery = fmt.Sprintf("id=%x", n.ID[:])
	return u.String()
}

func ParseNode(rawurl string) (*Node, error) {
	var (
		id NodeId
		ip net.IP
		tcpPort, udpPort uint64
	)
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("parse url err: %v", err)
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



// recoverNodeID computes the public key used to sign the
// given hash from the signature.
func recoverNodeID(buf []byte) (NodeId, error) {
	var id NodeId
	copy(id[:], buf)
	return id, nil
}
var lzcount = [256]int{
	8, 7, 6, 6, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}
// logdist returns the logarithmic distance between a and b, log2(a ^ b).
func logdist(a, b []byte) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += lzcount[x]
			break
		}
	}
	return len(a)*8 - lz
}

// distcmp compares the distances a->target and b->target.
// Returns -1 if a is closer to target, 1 if b is closer to target
// and 0 if they are equal.
func distcmp(target, a, b []byte) int {
	for i := range target {
		da := a[i] ^ target[i]
		db := b[i] ^ target[i]
		if da > db {
			return 1
		} else if da < db {
			return -1
		}
	}
	return 0
}
