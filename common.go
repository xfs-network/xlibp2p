package p2p

const (
	version1 uint8 = 1
)

const (
	typeHelloRequest uint8 = 0
	typeReHelloRequest uint8 = 1
	typePingMsg uint8 = 2
	typePongMsg uint8 = 3
)

func SendMsgData(p Peer, mType uint8, obj interface{}) error {
	return p.WriteMessageObj(mType,obj)
}