package bench

import (
	"fmt"
	"time"

	clientlib "github.com/scalog/scalog-client/lib"
)

type Bench struct {
	num    int32
	size   int32
	data   string
	client *clientlib.Client
}

func NewBench(num, size int32) (*Bench, error) {
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	b := &Bench{num: num, size: size}
	b.client = client
	b.data = string(make([]byte, size))
	return b, nil
}

func (b *Bench) Start() error {
	start := time.Now()
	for i := int32(0); i < b.num; i++ {
		_, err := b.client.Append(b.data)
		if err != nil {
			return err
		}
	}
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("%d Append operations of %d bytes each elapsed: %d ms\n", b.num, b.size, elapsed)
	return nil
}
