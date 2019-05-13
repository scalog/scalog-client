package bench

import (
	clientlib "github.com/scalog/scalog-client/lib"
	log "github.com/scalog/scalog/logger"
)

type Bench struct {
	num    uint32
	size   uint32
	client *clientlib.Client
	data   []byte
}

func NewBench(num, size uint32) *Bench {
	b := &Bench{num, size}
	b.client = NewClient()
	b.data = make([]byte, b.size)
	return b
}

func (b *bench) Start() (string, error) {
	for i := uint32(0); i < b.num; i++ {
		n, err := b.client.Append(b.data)
		if n != b.size {
			log.Errorf("Append returned length error: expect %d, get %d", b.size, n)
		}
	}
}
