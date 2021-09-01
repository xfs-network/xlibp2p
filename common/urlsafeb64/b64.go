package urlsafeb64

import (
	"encoding/base64"
	"strings"
)

func Encode(src []byte) string {
	sEnc := base64.StdEncoding.EncodeToString(src)
	sEnc = strings.ReplaceAll(sEnc, "+", "-")
	sEnc = strings.ReplaceAll(sEnc, "/", "_")
	sEnc = strings.ReplaceAll(sEnc, "=", "")
	return sEnc
}

func Decode(enc string) ([]byte, error) {
	encStr := enc
	encStr = strings.ReplaceAll(encStr, "-", "+")
	encStr = strings.ReplaceAll(encStr, "_", "/")
	return base64.RawStdEncoding.DecodeString(encStr)
}
