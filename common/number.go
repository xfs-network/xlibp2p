package common

import "encoding/json"

func Uint64s(t json.Number) (uint64, error) {
	number, err := t.Int64()
	if err != nil {
		return 0, err
	}
	return uint64(number), nil
}
