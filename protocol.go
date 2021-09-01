package p2p

type Protocol interface {
	Run(p Peer) error
}
