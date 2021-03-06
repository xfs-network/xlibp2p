package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"github.com/xfs-network/xlibp2p/discover"
	"github.com/xfs-network/xlibp2p/log"
	"github.com/xfs-network/xlibp2p/nat"
	"net"
	"sync"
	"time"
)

const (
	flagInbound = 1
	flagOutbound = 1 << 1
	flagStatic = 1 << 2
	flagDynamic = 1 << 3
)


type Server interface {
	Node() *discover.Node
	NodeId() discover.NodeId
	Peers() []Peer
	AddPeer(node *discover.Node)
	RemovePeer(node discover.NodeId)
	Bind(p Protocol)
	Start() error
	Stop()
}

// server manages all peer connections.
//
// The fields of Server are used as configuration parameters.
// You should set them before starting the Server. Fields may not be
// modified while the server is running.
type server struct {
	nodeId discover.NodeId
	node *discover.Node
	config Config
	mu     sync.Mutex
	running bool
	//protocols contains the protocols supported by the server.
	//Matching protocols are launched for each peer.
	protocols []Protocol
	close chan struct{}
	addpeer chan *peerConn
	addstatic chan *discover.Node
	rmstatic chan discover.NodeId
	delpeer chan Peer
	peers map[discover.NodeId]Peer
	table *discover.Table
	logger log.Logger
	lastLookup time.Time
}

// Config Background network service configuration
type Config struct {
	Nat nat.Mapper
	ListenAddr      string
	Key             *ecdsa.PrivateKey
	Discover bool
	NodeDBPath string
	StaticNodes     []*discover.Node
	BootstrapNodes []*discover.Node
	MaxPeers int
	Logger log.Logger
	Encoder encoder
}

// NewServer Creates background service object
func NewServer(config Config) Server {
	srv := &server{
		config:  config,
		logger: config.Logger,
	}
	if config.Logger == nil {
		srv.logger = log.DefaultLogger()
	}
	currentKey := srv.config.Key
	srv.nodeId = discover.PubKey2NodeId(currentKey.PublicKey)
	return srv
}

// Bind network protocol function
func (srv *server) Bind(p Protocol) {
	if srv.protocols == nil {
		srv.protocols = make([]Protocol, 0)
	}
	// Add network protocol
	srv.protocols = append(srv.protocols, p)
}

// Stop background network function
func (srv *server) Stop() {
	close(srv.close)
	srv.table.Close()
}

type udpcnn interface {
	LocalAddr() net.Addr
}

func (srv *server) listenUDP() (*discover.Table, udpcnn, error ) {
	addr, err := net.ResolveUDPAddr("udp", srv.config.ListenAddr)
	if err != nil {
		return nil, nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, nil, err
	}
	table, _ := discover.NewUDP(srv.config.Key, conn, srv.config.NodeDBPath, srv.config.Nat)
	return table, conn, nil
}

// Start start running the server.
func (srv *server) Start() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.running {
		return errors.New("server already running")
	}

	srv.running = true
	// Peer to peer session entity
	srv.addpeer = make(chan *peerConn)
	srv.addstatic = make(chan *discover.Node)
	srv.rmstatic = make(chan discover.NodeId)
	srv.delpeer = make(chan Peer)
	srv.close = make(chan struct{})
	var err error
	var uconn udpcnn = nil
	// launch node discovery and UDP listener
	if srv.config.Discover {
		srv.table, uconn, err = srv.listenUDP()
		if err != nil {
			return err
		}

	}
	dynPeers := srv.config.MaxPeers / 2
	if !srv.config.Discover {
		dynPeers = 0
	}
	dialer := newDialState(srv.config.StaticNodes, srv.table, dynPeers)
	// launch TCP listener to accept connection
	realaddr := uconn.LocalAddr().(*net.UDPAddr)
	if err = srv.listenAndServe(realaddr.Port); err != nil {
		return err
	}

	go srv.run(dialer)
	srv.running = true
	return nil
}

func (srv *server) run(dialer *dialstate) {
	srv.peers = make(map[discover.NodeId]Peer)
	tasks := make([]task, 0)
	pendingTasks := make([]task, 0)
	taskdone := make(chan task)
	delTask := func(t task) {
		for i := range tasks {
			if tasks[i] == t {
				tasks = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
	}

	scheduleTasks := func(new []task) {
		pt := append(pendingTasks, new...)
		start := 16 - len(tasks)
		if len(pt) < start {
			start = len(pt)
		}
		if start > 0 {
			tasks = append(tasks, pt[:start]...)
			for _, t := range pt[:start] {
				tt := t
				go func() {
					tt.Do(srv)
					taskdone <- tt
				}()
			}
			copy(pt, pt[start:])
			// pending tasks
			pendingTasks = pt[:len(pt)-start]
		}
	}
	for {
		now := time.Now()
		nt := dialer.newTasks(len(pendingTasks)+len(tasks), srv.peers, now)
		// schedule tasks
		scheduleTasks(nt)
		select {
		case n := <-srv.addstatic:
			dialer.addStatic(n)
		case n := <-srv.rmstatic:
			dialer.removeStatic(n)
			for k, v := range srv.peers {
				if bytes.Equal(k[:], n[:]) {
					v.Close()
				}
			}
			delete(srv.peers, n)
		// add peer
		case c := <-srv.addpeer:
			p := newPeer(c, srv.protocols, srv.config.Encoder)
			srv.peers[c.id] = p
			srv.logger.Infof("save peer id to peers: %s", c.id)
			go srv.runPeer(p)
		// task is done
		case t := <-taskdone:
			dialer.taskDone(t, now)
			delTask(t)
		// delete peer
		case p := <-srv.delpeer:
			pId := p.ID()
			delete(srv.peers, pId)
		}
	}
}

func (srv *server) runPeer(peer Peer) {
	peer.Run()
	srv.delpeer <- peer
}

func (srv *server) listenAndServe(realPort int) error {
	addr, err := net.ResolveTCPAddr("tcp", srv.config.ListenAddr)
	addr.Port = realPort
	ln, err := net.ListenTCP("tcp", addr)
	laddr := ln.Addr().(*net.TCPAddr)
	if err != nil {
		srv.logger.Errorf("p2p listen and serve on %s err: %v", laddr, err)
		return err
	}
	srv.logger.Infof("p2p listen and serve on %s", laddr)

	srv.node = discover.NewNode(addr.IP, uint16(addr.Port), uint16(addr.Port), srv.nodeId)
	srv.logger.Infof("p2p server node id: %s", srv.nodeId)
	go srv.listenLoop(ln)
	if !laddr.IP.IsLoopback() && srv.config.Nat != nil {
		//srv.loopWG.Add(1)
		go func() {
			srv.logger.Debugf("nat mapping \"xlibp2p server\" port: %d", laddr.Port)
			nat.Map(srv.config.Nat, srv.close, "tcp", laddr.Port, laddr.Port, "xlibp2p server")
			//srv.loopWG.Done()
		}()
	}
	return nil
}

// listenLoop runs in its own goroutine and accepts
// request of connections.
func (srv *server) listenLoop(ln net.Listener) {
	defer func() {
		if err := ln.Close(); err != nil {
			srv.logger.Errorln(err)
		}
	}()
	for {
		rw, err := ln.Accept()
		if err != nil {
			srv.logger.Errorf("p2p listenner accept err %v", err)
			return
		}
		c := srv.newPeerConn(rw, flagInbound, nil)
		go c.serve()
	}
}

func (srv *server) AddPeer(node *discover.Node) {
	srv.addstatic <- node
}

func (srv *server) Peers() []Peer {
	tmp := make([]Peer, 0)
	for _, v := range srv.peers {
		tmp = append(tmp, v)
	}
	return tmp
}

func (srv *server) RemovePeer(nId discover.NodeId) {
	srv.rmstatic <- nId
}

func (srv *server) NodeId() discover.NodeId {
	return srv.nodeId
}

func (srv *server) Node() *discover.Node {
	return srv.node
}

func (srv *server) newPeerConn(rw net.Conn, flag int, dst *discover.NodeId) *peerConn {
	pubKey := srv.config.Key.PublicKey
	mId := discover.PubKey2NodeId(pubKey)
	c := &peerConn{
		logger: srv.logger,
		self:    mId,
		flag: flag,
		server:  srv,
		key:     srv.config.Key,
		rw:      rw,
		version: version1,
	}
	if dst != nil {
		c.id = *dst
	}
	return c
}
