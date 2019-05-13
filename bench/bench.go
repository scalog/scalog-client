package bench

import (
	clientlib "github.com/scalog/scalog-client/lib"
	log "github.com/scalog/scalog/logger"
)

type Bench struct {
	num    int32
	size   int32
	client *clientlib.Client
	data   string
}

func NewBench(num, size int32) *Bench {
	b := &Bench{num: num, size: size}
	b.client = clientlib.NewClient()
	b.data = string(make([]byte, b.size))
	return b
}

func (b *Bench) Start() error {
	for i := int32(0); i < b.num; i++ {
		n, err := b.client.Append(b.data)
		if n != b.size {
			log.Printf("Append returned length error: expect %d, get %d", b.size, n)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
