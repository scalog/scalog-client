package test

import (
	clientlib "github.com/scalog/scalog-client/lib"
)

type Test struct {
	client *clientlib.Client
}

func NewTest() (*Test, error) {
	t := &Test{}
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	t.client = client
	return t, err
}

func (t *Test) Start() error {
	// TODO: implement the test
	return nil
}
