package p2p

import (
	"bytes"
	"container/heap"
	"crypto/rand"
	"net"
	"time"
	"xlibp2p/discover"
)
const (
	// This is the amount of time spent waiting in between
	// redialing a certain node.
	dialHistoryExpiration = 30 * time.Second

	// Discovery lookups are throttled and can only run
	// once every few seconds.
	lookupInterval = 4 * time.Second
)

type task interface {
	Do(srv *server)
}

type dialtask struct {
	flag int
	dest *discover.Node
}

func (t *dialtask) Do(srv *server) {
	tcpAddr := t.dest.TcpAddr()
	coon, err := net.Dial("tcp", tcpAddr.String())
	if err != nil {
		return
	}
	id := t.dest.ID
	c := srv.newPeerConn(coon, t.flag, &id)
	c.serve()
}
type discoverTask struct {
	bootstrap bool
	result  []*discover.Node
}


func (t *discoverTask) Do(srv *server) {
	if t.bootstrap {
		srv.table.Bootstrap(srv.config.BootstrapNodes)
		return
	}
	next := srv.lastLookup.Add(lookupInterval)
	if now := time.Now(); now.Before(next) {
		time.Sleep(next.Sub(now))
	}
	srv.lastLookup = time.Now()
	var target discover.NodeId
	_, _ = rand.Read(target[:])
	t.result = srv.table.Lookup(target)
}
type waitExpireTask struct {
	time.Duration
}

func (t waitExpireTask) Do(_ *server) {
	time.Sleep(t.Duration)
}

type dialHistory []pastDial

func (h dialHistory) min() pastDial {
	return h[0]
}
func (h *dialHistory) add(id discover.NodeId, exp time.Time) {
	heap.Push(h, pastDial{id, exp})
}
func (h dialHistory) contains(id discover.NodeId) bool {
	for _, v := range h {
		if bytes.Equal(v.id[:], id[:]) {
			return true
		}
	}
	return false
}
func (h *dialHistory) expire(now time.Time) {
	for h.Len() > 0 && h.min().exp.Before(now) {
		heap.Pop(h)
	}
}

func (h dialHistory) Len() int           { return len(h) }
func (h dialHistory) Less(i, j int) bool { return h[i].exp.Before(h[j].exp) }
func (h dialHistory) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *dialHistory) Push(x interface{}) {
	*h = append(*h, x.(pastDial))
}
func (h *dialHistory) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}


// pastDial is an entry in the dial history.
type pastDial struct {
	id  discover.NodeId
	exp time.Time
}


type dialstate struct {
	static map[discover.NodeId]*discover.Node
	ntab discoverTable
	maxDynDials int
	dialing map[discover.NodeId]int
	lookupBuf []*discover.Node
	lookupRunning bool
	bootstrapped  bool
	randomNodes []*discover.Node
	hist        *dialHistory
}
type discoverTable interface {
	Self() *discover.Node
	Close()
	Bootstrap([]*discover.Node)
	Lookup(target discover.NodeId) []*discover.Node
	ReadRandomNodes([]*discover.Node) int
}

func newDialState(static []*discover.Node, table discoverTable, maxdyn int) *dialstate {
	d := &dialstate{
		ntab: table,
		maxDynDials: maxdyn,
		static: make(map[discover.NodeId]*discover.Node),
		dialing: make(map[discover.NodeId]int),
		randomNodes: make([]*discover.Node, maxdyn/2),
		hist: new(dialHistory),
	}
	for _, a := range static {
		d.static[a.ID] = a
	}
	return d
}
func (d *dialstate) newTasks(nRunning int, peers map[discover.NodeId]Peer, now time.Time) []task {
	var tasks []task
	addDial := func(flag int, n *discover.Node) bool {
		//the connection established needn't to join the pool
		_, dialing := d.dialing[n.ID]
		if dialing ||  peers[n.ID] != nil || d.hist.contains(n.ID) {
			return false
		}
		d.dialing[n.ID] = flag
		tasks = append(tasks, &dialtask{
			flag: flag,
			dest:   n,
		})
		return true
	}
	needDynDials := d.maxDynDials
	for _,p := range peers {
		if p.Is(flagDynamic) {
			needDynDials -= 1
		}
	}
	for _,flag := range d.dialing {
		if flag&flagDynamic != 0 {
			needDynDials -= 1
		}
	}
	d.hist.expire(now)

	for _, n := range d.static {
		addDial(flagOutbound|flagStatic, n)
	}
	randomCandidates := needDynDials / 2
	if randomCandidates > 0 && d.bootstrapped {
		n := d.ntab.ReadRandomNodes(d.randomNodes)
		for i := 0; i < randomCandidates && i < n; i++ {
			if addDial(flagOutbound|flagDynamic, d.randomNodes[i]) {
				needDynDials--
			}
		}
	}
	i := 0
	for ; i < len(d.lookupBuf) && needDynDials > 0; i++ {
		if addDial(flagOutbound|flagDynamic, d.lookupBuf[i]) {
			needDynDials--
		}
	}
	d.lookupBuf = d.lookupBuf[:copy(d.lookupBuf, d.lookupBuf[i:])]
	if len(d.lookupBuf) < needDynDials && !d.lookupRunning {
		d.lookupRunning = true
		tasks = append(tasks, &discoverTask{bootstrap: !d.bootstrapped})
	}

	if nRunning == 0 && len(tasks) == 0 && d.hist.Len() > 0 {
		t := &waitExpireTask{d.hist.min().exp.Sub(now)}
		tasks = append(tasks, t)
	}
	return tasks
}


func (d *dialstate) taskDone(t task, now time.Time) {
	switch mt := t.(type) {
	case *discoverTask:
		if mt.bootstrap {
			d.bootstrapped = true
		}
		d.lookupRunning = false
		d.lookupBuf = append(d.lookupBuf, mt.result...)
	case *dialtask:
		d.hist.add(mt.dest.ID, now.Add(dialHistoryExpiration))
		delete(d.dialing, mt.dest.ID)
	}
}


