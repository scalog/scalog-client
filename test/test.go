package test

import (
	clientlib "github.com/scalog/scalog-client/lib"
)

type Bench struct {
	client *clientlib.Client
}

func NewBench() *Bench {
	b := &Bench{}
	b.client = clientlib.NewClient()
	return b
}

func (b *Bench) Start() error {
	// TODO: implement the test
	return nil
}
