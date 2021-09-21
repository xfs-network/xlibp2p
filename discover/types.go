package discover

import (
	"github.com/xfs-network/xlibp2p/crypto"
	"net"
	"time"
)

var maxNeighbors int = 1024


type ping struct {
	Version    int
	From, To   rpcEndpoint
	Expiration uint64
}

// pong is the reply to ping.
type pong struct {
	// This field should mirror the UDP envelope address
	// of the ping packet, which provides a way to discover the
	// the external address (after NAT).
	To rpcEndpoint
	Expiration uint64 // Absolute timestamp at which the packet becomes invalid.
}

// findnode is a query for nodes close to the given target.
type findnode struct {
	Target     NodeId // doesn't need to be an actual public key
	Expiration uint64
}

// reply to findnode
type neighbors struct {
	Nodes      []rpcNode
	Expiration uint64
}

type rpcNode struct {
	IP  net.IP // len 4 for IPv4 or 16 for IPv6
	UDP uint16 // for discovery protocol
	TCP uint16 // for RLPx protocol
	ID  NodeId
}

type rpcEndpoint struct {
	IP  net.IP // len 4 for IPv4 or 16 for IPv6
	UDP uint16 // for discovery protocol
	TCP uint16 // for RLPx protocol
}

func expired(ts uint64) bool {
	return time.Unix(int64(ts), 0).Before(time.Now())
}


func (req *ping) handle(t *udp, from *net.UDPAddr, fromID NodeId) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if req.Version != Version {
		return errBadVersion
	}
	_ = t.send(from, pongPacket, pong{
		To: makeEndpoint(from, req.From.TCP),
		Expiration: uint64(time.Now().Add(expiration).Unix()),
	})
	if !t.handleReply(fromID, pingPacket, req) {
		// Note: we're ignoring the provided IP address right now
		go func() {
			_, _ = t.bond(true, fromID, from, req.From.TCP)
		}()
	}
	return nil
}

func (req *pong) handle(t *udp, from *net.UDPAddr, fromID NodeId) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, pongPacket, req) {
		return errUnsolicitedReply
	}
	return nil
}

func (req *findnode) handle(t *udp, from *net.UDPAddr, fromID NodeId) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if t.db.node(fromID) == nil {
		// No bond exists, we don't process the packet. This prevents
		// an attack vector where the discovery protocol could be used
		// to amplify traffic in a DDOS attack. A malicious actor
		// would send a findnode request with the IP address and UDP
		// port of the target as the source address. The recipient of
		// the findnode packet would then send a neighbors packet
		// (which is a much bigger packet than findnode) to the victim.
		return errUnknownNode
	}
	target := crypto.ByteHash256(fromID[:])
	t.mu.Lock()
	closest := t.closest(target, bucketSize).entries
	t.mu.Unlock()

	p := neighbors{Expiration: uint64(time.Now().Add(expiration).Unix())}
	// Send neighbors in chunks with at most maxNeighbors per packet
	// to stay below the 1280 byte limit.
	for i, n := range closest {
		p.Nodes = append(p.Nodes, nodeToRPC(n))
		if len(p.Nodes) == maxNeighbors || i == len(closest)-1 {
			_ = t.send(from, neighborsPacket, p)
			p.Nodes = p.Nodes[:0]
		}
	}
	return nil
}

func (req *neighbors) handle(t *udp, from *net.UDPAddr, fromID NodeId) error {
	if expired(req.Expiration) {
		return errExpired
	}
	if !t.handleReply(fromID, neighborsPacket, req) {
		return errUnsolicitedReply
	}
	return nil
}

