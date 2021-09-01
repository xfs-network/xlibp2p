package discover

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTable_(t *testing.T) {

}

type testNet struct {
	t *testing.T
	ns []*Node
}
func (t *testNet) ping(id NodeId, addr *net.UDPAddr) error {
	return nil
}
func (t *testNet) waitping(NodeId) error{
	return nil
}
func (t *testNet) findnode(toid NodeId, addr *net.UDPAddr, target NodeId) ([]*Node, error){
	return nil,nil
}
func (t *testNet) close(){

}

var (
	tn = &testNet{
		ns: []*Node{
			newNode(
				net.IP{127,0,0,1},
				9001,
				9002,
				NodeId{0x01,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87}),
			newNode(
				net.IP{127,0,0,1},
				9003,
				9004,
				NodeId{0x02,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87}),
			newNode(
				net.IP{127,0,0,1},
				9005,
				9006,
				NodeId{0x03,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87}),
			newNode(
				net.IP{127,0,0,1},
				9007,
				9008,
				NodeId{0x04,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87}),
		},
	}
)

func TestTable_Lookup(t *testing.T) {
	selfId := NodeId{
		01,43,253,127,128,129,230,91,
	}
	addr := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9001,
	}
	tab := newTable(tn,selfId, addr,"./d0")
	defer tab.Close()
	find := tab.Lookup(tn.ns[0].ID)
	_=find
}
func TestTable_pingpong(t *testing.T) {
	selfId := NodeId{
		01,43,253,127,128,129,230,91,
	}
	addr := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9001,
	}
	targetId := NodeId{
		02,43,253,127,128,129,230,91,
	}
	target := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9002,
	}
	tab := newTable(tn, selfId, addr,"./d0")
	defer tab.Close()
	w := &bondproc{done: make(chan struct{})}
	tab.pingpong(w,true, targetId, target,0)
	if w.err != nil {
		t.Fatal(w.err)
	}
	gotNode := w.n
	if !bytes.Equal(gotNode.ID[:],targetId[:]) {
		t.Fatalf("got id: %s, want: %s", gotNode, targetId)
	}
	realNode := tab.db.node(gotNode.ID)
	if !bytes.Equal(realNode.ID[:],targetId[:]) {
		t.Fatalf("got id: %s, want: %s", realNode, targetId)
	}
}


func TestTable_bond(t *testing.T) {
	selfId := NodeId{
		01,43,253,127,128,129,230,91,
	}
	addr := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9001,
	}
	targetId := NodeId{
		02,43,253,127,128,129,230,91,
	}
	target := &net.UDPAddr{
		IP: net.IP{127,0,0,1},
		Port: 9002,
	}
	tab := newTable(tn, selfId, addr,"./d0")
	defer tab.Close()
	n, err := tab.bond(true, targetId, target,9093)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(n.ID[:], targetId[:]){
		t.Fatalf("got id: %s, want: %s", n.ID, targetId)
	}
	var got *Node
	for _,item := range tab.buckets {
		es := item.entries
		for _,en := range es {
			if bytes.Equal(en.ID[:], targetId[:]){
				got = en
			}
		}
	}
	if got == nil {
		t.Fatalf("not found by id: %s", targetId)
	}
}

func TestA(t *testing.T) {
	bondslots := make(chan struct{}, 16)
	for i := 0; i < cap(bondslots); i++ {
		bondslots <- struct{}{}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i:=0; i<18; i++ {
			go func() {
				<-bondslots
				time.Sleep(1 * time.Second)
				defer func() {bondslots<- struct{}{}}()
			}()
		}
	}()
	for {
		size := len(bondslots)
		t.Logf("bond slots len: %d", size)
		//if size == 0 {
		//	break
		//}
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
}