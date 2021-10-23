package hashfile

import (
	md52 "crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"hash/fnv"
)

func init() {
	SupportedHashes["md5"] = newMd5
	SupportedHashes["sha256"] = newSha256
	SupportedHashes["sha512"] = newSha512
	SupportedHashes["fnv1a"] = newFnv1a
}

// genericHash is our implementation of the Hash interface
type genericHash struct {
	h hash.Hash
}

func (m *genericHash) Write(p []byte) (n int, err error) {
	return m.h.Write(p)
}

func (m *genericHash) Sum() []byte {
	return m.h.Sum([]byte{})
}

func newMd5(interface{}) Hash {
	return &genericHash{h: md52.New()}
}

func newSha256(interface{}) Hash {
	return &genericHash{h: sha256.New()}
}

func newFnv1a(interface{}) Hash {
	return &genericHash{h: fnv.New128a()}
}

func newSha512(interface{}) Hash {
	return &genericHash{h: sha512.New()}
}
