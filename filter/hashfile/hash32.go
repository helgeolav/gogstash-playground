package hashfile

import (
	"hash"
	crc322 "hash/crc32"
)

func init() {
	SupportedHashes["crc32"] = newCrc32
}

type crc32 struct {
	crc32 hash.Hash32
}

func (c *crc32) Write(p []byte) (n int, err error) {
	return c.crc32.Write(p)
}

func (c *crc32) Sum() []byte {
	return c.crc32.Sum([]byte{})
}

// newCrc32 creates a new CRC32 hash
func newCrc32(interface{}) Hash {
	crc := &crc32{}
	crc.crc32 = crc322.NewIEEE()
	return crc
}
