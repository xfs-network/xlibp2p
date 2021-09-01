package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"xlibp2p/common"
)

func ECDSASign2Hex(hash []byte, prv *ecdsa.PrivateKey) (string, error) {
	sig, err := ECDSASign(hash, prv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}
func ECDSASign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, errors.New("data not hash")
	}
	pub := prv.PublicKey
	c := pub.Curve
	_ = c
	r, s, err := ecdsa.Sign(rand.Reader, prv, hash)
	if err != nil {
		return nil, err
	}
	inBuf := bytes.NewBuffer(nil)
	if err = common.BytesMixed(r.Bytes(), 1, inBuf); err != nil {
		return nil, err
	}
	if err = common.BytesMixed(s.Bytes(), 1, inBuf); err != nil {
		return nil, err
	}
	if err = common.BytesMixed(pub.X.Bytes(), 1, inBuf); err != nil {
		return nil, err
	}
	if err = common.BytesMixed(pub.Y.Bytes(), 1, inBuf); err != nil {
		return nil, err
	}
	inBs := inBuf.Bytes()
	outBuf := bytes.NewBuffer(nil)
	if err = common.BytesMixed(inBs, 1, outBuf); err != nil {
		return nil, err
	}
	return outBuf.Bytes(), err
}
func VerifySignatureFromHex(data []byte, sigHex string) bool {
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	return VerifySignature(data, sigBytes)
}
func VerifySignature(data []byte, sig []byte) bool {
	totalLen := sig[0]
	sigAll := sig[1 : totalLen+1]
	sigBuf := bytes.NewBuffer(sigAll)
	rBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	r := new(big.Int).SetBytes(rBytes)

	sBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	s := new(big.Int).SetBytes(sBytes)
	xBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	x := new(big.Int).SetBytes(xBytes)
	yBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	y := new(big.Int).SetBytes(yBytes)
	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return ecdsa.Verify(pub, data, r, s)
}

func VerifySignatureByPublic(data []byte, sig []byte, pub *ecdsa.PublicKey) bool {
	totalLen := sig[0]
	sigAll := sig[1 : totalLen+1]
	sigBuf := bytes.NewBuffer(sigAll)
	rBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	r := new(big.Int).SetBytes(rBytes)

	sBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return false
	}
	s := new(big.Int).SetBytes(sBytes)
	return ecdsa.Verify(pub, data, r, s)
}

func ParsePubKeyFromSignature(sig []byte) (ecdsa.PublicKey, error) {
	totalLen := sig[0]
	sigAll := sig[1 : totalLen+1]
	sigBuf := bytes.NewBuffer(sigAll)
	rBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	_ = new(big.Int).SetBytes(rBytes)
	sBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	_ = new(big.Int).SetBytes(sBytes)
	xBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	x := new(big.Int).SetBytes(xBytes)
	yBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	y := new(big.Int).SetBytes(yBytes)
	return ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}
