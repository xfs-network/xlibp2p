package p2p

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"xlibp2p/discover"
	"xlibp2p/log"
)

// Peer to peer connection session
type peerConn struct {
	logger log.Logger
	inbound         bool
	id              discover.NodeId
	self            discover.NodeId
	server          *server
	key             *ecdsa.PrivateKey
	rw              net.Conn
	version         uint8
	handshakeStatus int
	flag int
}

func (c *peerConn) serve() {
	// Get the address and port number of the client
	fromAddr := c.rw.RemoteAddr()
	inbound := c.flag & flagInbound != 0
	if inbound {
		if err := c.serverHandshake(); err != nil {
			c.logger.Warnf("handshake error from %s: %v", fromAddr, err)
			c.close()
			return
		}
	} else {
		if err := c.clientHandshake(); err != nil {
			c.logger.Warnf("handshake error from %s: %v", fromAddr, err)
			c.close()
			return
		}
	}
	c.logger.Infof("p2p handshake success by %s", fromAddr)
	c.server.addpeer <- c
}

//Client handshake sending method
func (c *peerConn) clientHandshake() error {

	// Whether the handshake status is based on handshake
	if c.handshakeCompiled() {
		return nil
	}
	request := &helloRequestMsg{
		version:   c.version,
		id:        c.self,
		receiveId: c.id,
	}
	c.logger.Debugf("send hello request version: %d, id: %s, to receiveId: %s", c.version,c.self, c.id)
	_, err := c.rw.Write(request.marshal())
	if err != nil {
		return err
	}
	hello, err := c.readHelloReRequestMsg()

	if err != nil {
		return err
	}
	c.logger.Debugf("receive hello request reply from node: %s, by version: %d", hello.id, c.version)
	if hello.version != c.version {
		return fmt.Errorf("handshake check err, got version: %d, want version: %d",
			hello.version, c.version)
	}
	gotId := hello.receiveId
	wantId := c.self
	if !bytes.Equal(gotId[:], wantId[:]) {
		return fmt.Errorf("handshake check err got my name: 0x%x, my real name: 0x%x",
			gotId, wantId)
	}
	c.handshakeStatus = 1
	return nil
}

// Service handshake response method
func (c *peerConn) serverHandshake() error {
	// Whether the handshake status is based on handshake
	if c.handshakeCompiled() {
		return nil
	}

	// Read reply data
	hello, err := c.readHelloRequestMsg()
	if err != nil {
		return err
	}
	c.logger.Debugf("receive handshake request by nodeId %s, version: %d",hello.id, hello.version)
	if hello.version != c.version {
		return fmt.Errorf("handshake check err, got version: %d, want version: %d",
			hello.version, c.version)
	}
	gotId := hello.receiveId
	wantId := c.self
	if !bytes.Equal(gotId[:], wantId[:])  {
		return fmt.Errorf("handshake check err got my name: 0x%x, my real name: 0x%x",
			gotId, wantId)
	}
	c.id = hello.id

	reply := &helloReRequestMsg{
		id:        c.self,
		receiveId: hello.id,
		version:   c.version,
	}
	c.logger.Debugf("send handshake reply to nodeId %s", reply.receiveId)
	if _, err = c.rw.Write(reply.marshal()); err != nil {
		return err
	}
	return nil
}

// Read reply message
func (c *peerConn) readHelloReRequestMsg() (*helloReRequestMsg, error) {
	msg, err := c.readMessage()
	if err != nil {
		return nil, err
	}
	if msg.Type() != typeReHelloRequest {
		return nil, err
	}
	nMsg := new(helloReRequestMsg)
	raw, _ := ioutil.ReadAll(msg.RawReader())
	if !nMsg.unmarshal(raw) {
		return nil, errors.New("parse hello request err")
	}
	return nMsg, nil
}

// Read peer session messages
func (c *peerConn) readHelloRequestMsg() (*helloRequestMsg, error) {
	msg, err := c.readMessage()
	if err != nil {
		return nil, err
	}
	if msg.Type() != typeHelloRequest {
		return nil, err
	}
	nMsg := new(helloRequestMsg)
	raw, _ := ioutil.ReadAll(msg.RawReader())
	if !nMsg.unmarshal(raw) {
		return nil, errors.New("parse hello request err")
	}
	return nMsg, nil
}

// Write peer session messages
func (c *peerConn) writeMessage(mType uint8, data []byte) error {
	cLen := len(data)
	val := make([]byte, cLen+4)
	binary.LittleEndian.PutUint32(val, uint32(cLen))
	copy(val[4:], data)
	msg := []byte{c.version, mType}
	msg = append(msg, val...)
	_, err := c.rw.Write(msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *peerConn) readMessage() (MessageReader, error) {
	return ReadMessage(c.rw)
}

func (c *peerConn) close() {
	if err := c.rw.Close(); err != nil {
		c.logger.Errorln(err)
	}
}

func (c *peerConn) handshakeCompiled() bool {
	return c.handshakeStatus == 1
}
