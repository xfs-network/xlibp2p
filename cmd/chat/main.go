package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	p2p "xlibp2p"
	"xlibp2p/crypto"
	"xlibp2p/discover"
	"xlibp2p/nat"
)

type chatProtocol struct {
	ps map[discover.NodeId]p2p.Peer
}

func (cp *chatProtocol) Run(p p2p.Peer) error {
	if cp.ps == nil {
		cp.ps = make(map[discover.NodeId]p2p.Peer)
	}
	cp.ps[p.ID()] = p
	defer delete(cp.ps, p.ID())
	out:
	for {
		select {
		case <-p.QuitCh():
			break out
		default:
		}
		if err := cp.handleMsg(p); err != nil {
			return err
		}
	}
	return nil
}


func (cp *chatProtocol) handleMsg(p p2p.Peer) error {
	ch := p.GetProtocolMsgCh()
	select {
	case mr := <-ch:
		if mr.Type() == 4 {
			bs,_ := mr.ReadAll()
			nId := p.ID()
			fmt.Printf("<(%x...%x): %s\n",nId[:3],nId[len(nId)-3:], string(bs))
		}
	}
	return nil
}

func (cp *chatProtocol) sendMessage(txt string) {
	for _, p := range cp.ps {
		err := p.WriteMessage(4, []byte(txt))
		if err != nil {
			continue
		}
	}
}

var (
	addr string
	datadir string
	help bool
	bootstrap string
	static string
	maxPeers int
)

func init() {
	flag.StringVar(&addr, "addr", ":", "set listen address")
	flag.StringVar(&datadir, "datadir", "", "set data dir path (default: random folder in current path)")
	flag.StringVar(&bootstrap, "bootstrap", "", "set bootstrap nodes")
	flag.StringVar(&static, "static", "", "set static nodes")
	flag.IntVar(&maxPeers, "maxpeers", 10,"set bootstrap nodes")
	flag.BoolVar(&help, "help", false, "this help")
}

func randomStr(len int) string {
	cs := "0123456789abcdefghijklmnopqrstuvwxyz"
	buf := ""
	for i:=0; i < len; i++ {
		index := rand.Intn(34) + 1
		buf += string(cs[index])
	}
	return buf
}
func randomDataDir() string {
	var currentPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		currentPath = path.Dir(filename)
	}
	p := path.Join(currentPath, randomStr(8))
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		panic(err)
	}
	return p
}
func resolveNodeUris(s string) []*discover.Node {
	if s == "" {
		return nil
	}
	sn := strings.Split(s, ",")
	buf := make([]*discover.Node, 0)
	for _,item := range sn {
		node, err := discover.ParseNode(item)
		if err != nil {
			continue
		}
		buf = append(buf, node)
	}
	return buf
}
func input(fn func(string)) {
	var (
		r = bufio.NewReader(os.Stdin)

	)
	r = bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		buf, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}
		line := string(buf)
		if len(line) == 0 {
			continue
		}
		fn(line)
	}
}
func chat(cp *chatProtocol, txt string) {
	if txt == "/peers" {
		for pId,_ := range cp.ps {
			fmt.Printf("peer id: %s\n", pId)
		}
		return
	}
	cp.sendMessage(txt)
}
func main() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	datadir = randomDataDir()
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.InfoLevel)
	key := crypto.MustGenPrvKey()
	bootNodes := resolveNodeUris(bootstrap)
	ss := resolveNodeUris(static)
	srv := p2p.NewServer(p2p.Config{
		Nat: nat.Any(),
		StaticNodes: ss,
		ListenAddr: addr,
		Key: key,
		BootstrapNodes: bootNodes,
		Discover: true,
		NodeDBPath: datadir,
		MaxPeers: maxPeers,
		Logger: logger,
	})
	cp := &chatProtocol{}
	srv.Bind(cp)
	if err := srv.Start(); err != nil {
		panic(err)
	}
	go input(func(s string) {
		chat(cp, s)
	})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	srv.Stop()
	if err := os.RemoveAll(datadir); err != nil {
		panic(err)
	}
	if err := os.Stdin.Close(); err != nil {
		panic(err)
	}
	println()
}
