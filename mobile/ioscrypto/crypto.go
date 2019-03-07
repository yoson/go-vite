package ioscrypto

import (
	"github.com/vitelabs/go-vite/crypto"
	"github.com/vitelabs/go-vite/crypto/ed25519"
)

func AesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	return crypto.AesCTRXOR(key, inText, iv)
}

func Hash256(data []byte) []byte {
	return crypto.Hash256(data)
}

func Hash(size int, data []byte) []byte {
	return crypto.Hash(size, data)
}

type Ed25519KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

type SignDataResult struct {
	PublicKey []byte
	Message   []byte
	Signature []byte
}

func GenerateEd25519KeyPair(seed []byte) (p *Ed25519KeyPair, _ error) {
	var s [32]byte
	copy(s[:], seed[:])
	publicKey, privateKey, err := ed25519.GenerateKeyFromD(s)
	if err != nil {
		return nil, err
	}
	return &Ed25519KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

func SignData(priv []byte, message []byte) *SignDataResult {
	var a ed25519.PrivateKey = priv
	signature := ed25519.Sign(a, message)

	return &SignDataResult{
		PublicKey: a.PubByte(),
		Message:   message,
		Signature: signature,
	}
}

func Ed25519PubToCurve25519(ed25519Pub []byte) []byte {
	var ep ed25519.PublicKey
	ep = ed25519Pub
	return ep.ToX25519Pk()
}

func Ed25519PrivToCurve25519(ed25519Priv []byte) []byte {
	var ep ed25519.PrivateKey
	ep = ed25519Priv
	return ep.ToX25519Sk()
}

func X25519ComputeSecret(private []byte, peersPublic []byte) ([]byte, error) {
	return crypto.X25519ComputeSecret(private, peersPublic)
}

func VerifySignature(pub, message, signData []byte) (bool, error) {
	return crypto.VerifySig(pub, message, signData)
}
