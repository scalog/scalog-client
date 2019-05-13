package it

import (
	clientlib "github.com/scalog/scalog-client/lib"
)

type It struct {
	client *clientlib.Client
}

func NewIt() *It {
	it := &It{}
	it.client = clientlib.NewClient()
	return it
}

func (it *It) Start() error {
	// TODO: implement the interactive client
	return nil
}
