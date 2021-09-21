package discover

import (
	"bytes"
	"encoding/binary"
	"github.com/xfs-network/xlibp2p/common/rawencode"
	"github.com/xfs-network/xlibp2p/crypto"
	"github.com/xfs-network/xlibp2p/storage/badger"
	"sync"
	"time"
)

type nodeDB struct {
	storage *badger.Storage
	version uint32
	self NodeId
	quit chan struct{}
	runner sync.Once
	seeder badger.Iterator
}
var (
	nodeDBNilNodeID      = NodeId{}       // Special node ID to use as a nil element.
	nodeDBNodeExpiration =  24 * time.Hour // Time after which an unseen node should be dropped.
	nodeDBCleanupCycle   = time.Hour    // Time period for running the expiration task.
	nodeDBItemPrefix = []byte("n:")
	nodeDBDiscoverRoot      = ":discover"
	nodeDBDiscoverPing      = nodeDBDiscoverRoot + ":lastping"
	nodeDBDiscoverPong      = nodeDBDiscoverRoot + ":lastpong"
	nodeDBDiscoverFindFails = nodeDBDiscoverRoot + ":findfail"
)

func newNodeDB(path string, version uint32, self NodeId) (*nodeDB, error) {
	db := &nodeDB{
		version: version,
		self: self,
		quit: make(chan struct{}),
	}
	var err error
	db.storage, err = badger.NewByVersion(path, version)
	return db, err
}

func (db *nodeDB) close() {
	if db.seeder != nil {
		db.seeder.Close()
	}
	err := db.storage.Close()
	if err != nil {
		panic(err)
	}
	close(db.quit)
}
func makeKey(id NodeId, field string) []byte {
	if bytes.Equal(id[:], nodeDBNilNodeID[:]) {
		return []byte(field)
	}
	return append(nodeDBItemPrefix, append(id[:], []byte(field)...)...)
}

func splitKey(key []byte) (id NodeId, field string) {
	// If the key is not of a node, return it plainly
	if !bytes.HasPrefix(key, nodeDBItemPrefix) {
		return NodeId{}, string(key)
	}
	// Otherwise split the id and field
	item := key[len(nodeDBItemPrefix):]
	copy(id[:], item[:len(id)])
	field = string(item[len(id):])

	return id, field
}
func (db *nodeDB) node(id NodeId) *Node {
	blob,err := db.storage.GetData(makeKey(id, nodeDBDiscoverRoot))
	if err != nil {
		return nil
	}
	node := new(Node)
	if err := rawencode.Decode(blob, node); err != nil {
		return nil
	}
	node.Hash = crypto.ByteHash256(node.ID[:])
	return node
}
// updateNode inserts - potentially overwriting - a node into the peer database.
func (db *nodeDB) updateNode(node *Node) error {
	blob, err := rawencode.Encode(node)
	if err != nil {
		return err
	}
	return db.storage.SetData(makeKey(node.ID,nodeDBDiscoverRoot),blob)
}

// deleteNode deletes all information/keys associated with a node.
func (db *nodeDB) deleteNode(id NodeId) error {
	it := db.storage.NewIterator()
	defer it.Close()
	for it.Next() {
		k := it.Key()
		if !bytes.HasPrefix(k, nodeDBItemPrefix) {
			continue
		}
		nid, _ := splitKey(k)
		if !bytes.Equal(nid[:],id[:]){
			continue
		}
		err := db.storage.DelData(k)
		if err != nil {
			continue
		}
	}
	return nil
}

func (db *nodeDB) fetchInt64(key []byte) int64 {
	blob, err := db.storage.GetData(key)
	if err != nil {
		return 0
	}
	val, read := binary.Varint(blob)
	if read <= 0 {
		return 0
	}
	return val
}

func (db *nodeDB) storeInt64(key []byte, n int64) error {
	blob := make([]byte, binary.MaxVarintLen64)
	blob = blob[:binary.PutVarint(blob, n)]
	return db.storage.SetData(key, blob)
}
func (db *nodeDB) findFails(id NodeId) int {
	return int(db.fetchInt64(makeKey(id, nodeDBDiscoverFindFails)))
}

// updateFindFails updates the number of findnode failures since bonding.
func (db *nodeDB) updateFindFails(id NodeId, fails int) error {
	return db.storeInt64(makeKey(id, nodeDBDiscoverFindFails), int64(fails))
}

// lastPing retrieves the time of the last ping packet send to a remote node,
// requesting binding.
func (db *nodeDB) lastPing(id NodeId) time.Time {
	return time.Unix(db.fetchInt64(makeKey(id, nodeDBDiscoverPing)), 0)
}

// updateLastPing updates the last time we tried contacting a remote node.
func (db *nodeDB) updateLastPing(id NodeId, instance time.Time) error {
	return db.storeInt64(makeKey(id, nodeDBDiscoverPing), instance.Unix())
}

// lastPong retrieves the time of the last successful contact from remote node.
func (db *nodeDB) lastPong(id NodeId) time.Time {
	return time.Unix(db.fetchInt64(makeKey(id, nodeDBDiscoverPong)), 0)
}

// updateLastPong updates the last time a remote node successfully contacted.
func (db *nodeDB) updateLastPong(id NodeId, instance time.Time) error {
	return db.storeInt64(makeKey(id, nodeDBDiscoverPong), instance.Unix())
}


// expireNodes iterates over the database and deletes all nodes that have not
// been seen (i.e. received a pong from) for some alloted time.
func (db *nodeDB) expireNodes() error {
	threshold := time.Now().Add(-nodeDBNodeExpiration)
	//db.Logger.Debugf("expired inspect threshold: %s", threshold)
	return db.storage.ForeachData(func(k []byte, v []byte) error {
		id, field := splitKey(k)
		// Skip the item if not a discovery node
		if field != nodeDBDiscoverRoot {
			return nil
		}
		// Skip the node if not expired yet (and not self)
		if bytes.Compare(id[:], db.self[:]) != 0{
			if seen := db.lastPong(id); seen.After(threshold) {
				//db.Logger.Debugf("expired inspect: %s: live", id)
				return nil
			}
		}
		//db.Logger.Debugf("expired inspect: %s: overdue", id)
		if err := db.deleteNode(id); err != nil {
			//db.Logger.Warnln("del expireNodes err", err)
		}
		return nil
	})
}
func (db *nodeDB) ensureExpire() {
	db.runner.Do(func() { go db.expire() })
}
// expirer should be started in a go routine, and is responsible for looping ad
// infinitum and dropping stale data from the database.
func (db *nodeDB) expire() {
	tick := time.Tick(nodeDBCleanupCycle)
	for {
		select {
		case <-tick:
			if err := db.expireNodes(); err != nil {
				//db.Logger.Infof("Failed to expire nodedb items: %v", err)
			}
		case <-db.quit:
			return
		}
	}
}

func (db *nodeDB) querySeeds(n int) []*Node {
	// Create a new seed iterator if none exists
	if db.seeder == nil {
		db.seeder = db.storage.NewIterator()
	}
	// Iterate over the nodes and find suitable seeds
	nodes := make([]*Node, 0, n)
	for len(nodes) < n && db.seeder.Next() {
		// Iterate until a discovery node is found
		id, field := splitKey(db.seeder.Key())
		if field != nodeDBDiscoverRoot {
			continue
		}
		// Dump it if its a self reference
		if bytes.Compare(id[:], db.self[:]) == 0 {
			if err := db.deleteNode(id); err != nil {
				//db.Logger.Warnln("deleteNode err", err)
				continue
			}
			continue
		}
		// Load it as a potential seed
		if node := db.node(id); node != nil {
			nodes = append(nodes, node)
		}
	}
	// Release the iterator if we reached the end
	if len(nodes) == 0 {
		db.seeder.Close()
		db.seeder = nil
	}
	return nodes
}

