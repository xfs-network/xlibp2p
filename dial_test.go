package p2p

import (
	"crypto/ecdsa"
	"net"
	"testing"
	"time"
	"xlibp2p/crypto"
	"xlibp2p/discover"
)

var boots = []string{
	"xfsnode://127.0.0.1:9092/?id=ff929a9b96a52935b3b65808c73e645b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f",
	"xfsnode://127.0.0.1:9092/?id=ff729a9b96a52935b3b65808c73e649b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f",
	"xfsnode://127.0.0.1:9092/?id=ff929a9b96a52935b3b65809c73e645b7bbf13dc53f6e0ce140079f106b67fa48d4690500db818c28618f70a04125d6707e38578904c7187fd755b8184f0375f",
}

type testTable struct {
	key *ecdsa.PrivateKey
	self *discover.Node
	t *testing.T
}
func (t *testTable) Self() *discover.Node {
	return t.self
}

func (t *testTable) Close() {

}

func (t *testTable) Bootstrap(ns []*discover.Node) {

}

func (t *testTable) Lookup(nid discover.NodeId) []*discover.Node {
	return nil
}

func (t *testTable) ReadRandomNodes([]*discover.Node) int {
	return 0
}

func newTestTable(t *testing.T, key *ecdsa.PrivateKey) *testTable {
	if key == nil {
		t.Fatal("key not set")
	}
	id := discover.PubKey2NodeId(key.PublicKey)
	node := discover.NewNode(net.IP{127,0,0,1}.To4(),1024,1024, id)
	return &testTable{
		key: key,
		t: t,
		self: node,
	}
}

func Test_dialtask_newDialState(t *testing.T) {
	key := crypto.MustGenPrvKey()
	bootNs := make([]*discover.Node, 0)
	for _, nRaw := range boots {
		n, err := discover.ParseNode(nRaw)
		if err != nil {
			t.Fatal(err)
		}
		bootNs = append(bootNs, n)
	}
	dynPeers := 10 / 2
	ds := newDialState(bootNs,newTestTable(t,key),dynPeers)
	ps := make(map[discover.NodeId]Peer)
	for {
		now := time.Now()
		ts := ds.newTasks(1,ps,now)
		for _, tt := range ts {
			tt.Do(nil)
			ds.taskDone(tt,now)
		}
		time.Sleep(10 * time.Second)
	}


}
