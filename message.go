package p2p

import (
	"bytes"
	"encoding/binary"
	"io"
	"xlibp2p/discover"
)

const headerLen = 6
//MessageReader interface defines type of message and reading methods,
//messageReader implements this interface.
type MessageReader interface {
	Type() uint8                // message type
	Read(p []byte) (int, error) // read given data to buffer
	ReadAll() ([]byte, error)   // read all of messages
	RawReader() io.Reader       // read raw messages(include version and mType fields)
	DataReader() io.Reader      // read data of message
}

//messageReader defines the message struct
type messageReader struct {
	version uint8
	mType   uint8
	raw     io.Reader
	data    io.Reader
}

// Type returns message type
func (m *messageReader) Type() uint8 {
	return m.mType
}

// Read reads the given data in the data part from messageReader
func (m *messageReader) Read(p []byte) (int, error) {
	return m.data.Read(p)
}

// ReadAll reads all of data in the message
func (m *messageReader) ReadAll() ([]byte, error) {
	return io.ReadAll(m.data)
}

// RawReader reads the whole package from other peer.
func (m *messageReader) RawReader() io.Reader {
	return m.raw
}

// DataReader  reads the data from the messageReader
func (m *messageReader) DataReader() io.Reader {
	return m.data
}

// ReadMessage reads message from other peer and returns MessageReader by header of message.
// message = version(1byte)+type(1byte)+length(4byte)+data
func ReadMessage(reader io.Reader) (MessageReader, error) {
	mBuffer := bytes.NewBuffer(nil)

	//vertion and type
	for mBuffer.Len() < 6 {
		b := make([]byte, 1)
		if _, err := reader.Read(b); err != nil {
			return nil, err
		}
		mBuffer.Write(b)
	}
	header := mBuffer.Bytes()
	//length of data in message.4 bytes stored by LittleEndian model.
	n := binary.LittleEndian.Uint32(header[2:])

	for mBuffer.Len() < 6+int(n) {
		b := make([]byte, 1)

		if _, err := reader.Read(b); err != nil {
			return nil, err
		}

		mBuffer.Write(b)
	}

	data := mBuffer.Bytes()

	return &messageReader{
		raw:   mBuffer,
		mType: data[1],
		data:  bytes.NewReader(data[6:]),
	}, nil
}

type helloRequestMsg struct {
	raw       []byte
	version   uint8
	id        discover.NodeId
	receiveId discover.NodeId
}

func (m *helloRequestMsg) marshal() []byte {
	if m.raw != nil {
		return m.raw
	}
	cLen := len(m.id) + len(m.receiveId)
	val := make([]byte, cLen+4)
	binary.LittleEndian.PutUint32(val, uint32(cLen))
	copy(val[4:], append(m.id[:], m.receiveId[:]...))
	base := []byte{m.version, typeHelloRequest}
	base = append(base, val...)
	return base
}

func (m *helloRequestMsg) unmarshal(data []byte) bool {
	m.raw = data
	m.version = data[0]
	if data[1] != typeHelloRequest {
		return false
	}
	cLen := binary.LittleEndian.Uint32(data[2:headerLen])
	body := data[headerLen: headerLen + cLen]
	copy(m.id[:], body[:len(m.id)])
	copy(m.receiveId[:], body[len(m.id):])
	return true
}

type helloReRequestMsg struct {
	raw       []byte
	version   uint8
	id        discover.NodeId
	receiveId discover.NodeId
}

func (m *helloReRequestMsg) marshal() []byte {
	if m.raw != nil {
		return m.raw
	}
	cLen := len(m.id) + len(m.receiveId)
	val := make([]byte, cLen+4)
	binary.LittleEndian.PutUint32(val, uint32(cLen))
	copy(val[4:], append(m.id[:], m.receiveId[:]...))
	base := []byte{m.version, typeReHelloRequest}
	base = append(base, val...)
	return base
}

func (m *helloReRequestMsg) unmarshal(data []byte) bool {
	m.raw = data
	m.version = data[0]
	if data[1] != typeReHelloRequest {
		return false
	}
	cLen := binary.LittleEndian.Uint32(data[2:headerLen])
	body := data[headerLen: headerLen + cLen]
	copy(m.id[:], body[:len(m.id)])
	copy(m.receiveId[:], body[len(m.id):])
	return true
}
