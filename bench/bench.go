package bench

import (
	"fmt"

	clientlib "github.com/scalog/scalog-client/lib"
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
		if err != nil {
			return err
		}
		if n != b.size {
			return fmt.Errorf("Append returned length error: expect %d, get %d", b.size, n)
		}
	}
	return nil
}
