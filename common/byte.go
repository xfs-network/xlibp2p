package common

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
)


func Decode16Byte(str string) []byte {
	result, _ := hex.DecodeString(str)
	return result
}

func Encode16Byte(byteVal []byte) string {
	result := hex.EncodeToString(byteVal)
	return result
}
func BytesMixed(src []byte, lenBits int, buffer *bytes.Buffer) error {
	srcLen := len(src)
	if uint32(srcLen) > uint32(math.MaxUint32) {
		return errors.New("data to long")
	}
	var lenBuf [4]byte
	lenBuf[0] = uint8(srcLen & 0xff)
	lenBuf[1] = uint8((srcLen & 0xff00) >> 8)
	lenBuf[2] = uint8((srcLen & 0xff0000) >> 16)
	lenBuf[3] = uint8((srcLen & 0xff000000) >> 32)
	buffer.Write(lenBuf[0:lenBits])
	buffer.Write(src)
	return nil
}

func ReadMixedBytes(buf *bytes.Buffer) ([]byte, error) {
	dataLenB, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	dataLen := int(dataLenB)
	var dst = make([]byte, dataLen)
	_, err = buf.Read(dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}
