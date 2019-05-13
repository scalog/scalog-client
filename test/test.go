package test

import (
	clientlib "github.com/scalog/scalog-client/lib"
)

type Test struct {
	client *clientlib.Client
}

func NewTest() *Test {
	t := &Test{}
	t.client = clientlib.NewClient()
	return t
}

func (t *Test) Start() error {
	// TODO: implement the test
	return nil
}
