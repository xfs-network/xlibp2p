package discover

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)
var (
	nodes = [4]*Node{
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
	}
)

func TestNodeDB_makeKey(t *testing.T) {
	var (
		nid = NodeId{
			0x01,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		}
		wantField = nodeDBDiscoverRoot
		offset = 0
	)
	got := makeKey(nid, wantField)
	t.Logf("key: %x", got)
	gotPrefix := got[:len(nodeDBItemPrefix)]
	offset += len(gotPrefix)
	if !bytes.Equal(gotPrefix, nodeDBItemPrefix) {
		t.Fatalf("got prefix: %s, want: %s", gotPrefix, nodeDBItemPrefix)
	}
	gotId := got[offset:offset+len(nid)]
	offset += len(gotId)
	if !bytes.Equal(gotId, nid[:]) {
		t.Fatalf("got id: %x, want: %s", gotId, nid)
	}
	gotField := got[offset:offset+len(wantField)]
	offset += len(gotField)
	if !bytes.Equal(gotField, []byte(wantField)) {
		t.Fatalf("got field: %s, want: %s", gotField, wantField)
	}
	if offset != len(got) {
		t.Fatalf("got len: %d, want: %d", len(got), offset)
	}
}

func TestNodeDB_splitKey(t *testing.T) {
	var (
		nid = NodeId{
			0x01,0x74,0x93,0x3a,0xe7,0x21,0x87,0x87,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
			0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		}
		wantField = nodeDBDiscoverRoot
	)
	key := makeKey(nid, wantField)
	t.Logf("key: %x", key)
	gotId, gotField := splitKey(key)
	if !bytes.Equal(gotId[:], nid[:]) {
		t.Fatalf("got id: %s, want: %s", gotId, nid)
	}
	if gotField != wantField {
		t.Fatalf("got field: %s, want: %s", gotField, wantField)
	}
}

func TestNodeDB_node(t *testing.T) {
	var (
		err error = nil
	)
	db, err := newNodeDB("./d0",Version, NodeId{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	for i,n := range nodes {
		if err = db.updateNode(n); err != nil {
			t.Fatalf("node[%d], %s", i, err)
		}
	}
	for i,item := range nodes {
		n := db.node(item.ID)
		if !bytes.Equal(n.ID[:], item.ID[:]){
			t.Fatalf("node[%d], got id %s, want: %s", i, n.ID, item.ID)
		}
	}
}

func TestNodeDB_querySeeds(t *testing.T) {
	var (
		err error = nil
	)
	db, err := newNodeDB("./d0",Version, NodeId{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	ns := db.querySeeds(len(nodes) / 2)
	ns2:= db.querySeeds(len(nodes) * 2)
	gotAll := append(ns, ns2...)
	if len(gotAll) > len(nodes) {
		t.Fatalf("got len: %d > %d", len(gotAll), len(nodes))
	}
}

func TestNodeDB_view(t *testing.T) {
	var (
		err error = nil
	)
	db, err := newNodeDB("./d0", Version, NodeId{})
	if err != nil {
		t.Fatal(err)
	}
	it := db.storage.NewIterator()
	defer func() {
		it.Close()
		db.close()
	}()
	i := 0
	for it.Next() {
		k := it.Key()
		id, field := splitKey(k)
		if field != nodeDBDiscoverRoot {
			continue
		}
		n := db.node(id)
		t.Logf("node[%d]: %s", i, n)
		i += 1
	}
}
func TestNodeDB_deleteNode(t *testing.T)  {
	var (
		err error = nil
	)
	db, err := newNodeDB("./d0",Version, NodeId{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	_ = db.deleteNode(nodes[1].ID)
}
func TestNodeDB_ensureExpire(t *testing.T) {
	var (
		err error = nil
	)
	db, err := newNodeDB("./d0",Version, NodeId{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	startTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(10 * time.Second)
		out:
		for {
			endTime := time.Now()
			if endTime.Sub(startTime) > 120 * time.Second {
				break
			}
			select {
			case <-tick:
				if err = db.updateLastPong(nodes[1].ID,time.Now()); err != nil {
					break out
				}
			}
		}
		t.Logf("exit update runner")
	}()
	db.ensureExpire()
	wg.Wait()
	n := db.node(nodes[1].ID)
	if n == nil {
		t.Fatal("expected value not met, nodes[1] is nil")
	}
}