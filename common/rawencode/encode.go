package rawencode

import "encoding/json"

type RawEncoder interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

func Encode(en interface{}) ([]byte, error) {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Marshal(en)
	}
	return obj.Encode()
}

func EncodeByLen(en interface{}) ([]byte, error) {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Marshal(en)
	}
	return obj.Encode()
}

func Decode(bs []byte, en interface{}) error {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Unmarshal(bs, en)
	}
	return obj.Decode(bs)
}
