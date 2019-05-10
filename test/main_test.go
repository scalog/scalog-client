package test

import (
	"fmt"
	"testing"

	"github.com/scalog/scalog-client/client"
)

func TestClientAppend(t *testing.T) {
	client := client.NewClient()
	resp, err := client.Append("Hello, World!")
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println(fmt.Sprintf("Append responded with global sequence number: %d", resp))
}

func TestClientSubscribe(t *testing.T) {
	client := client.NewClient()
	c := client.Subscribe(1)
	appendResp, err := client.Append("Hello, World!")
	fmt.Println(fmt.Sprintf("Append responded with global sequence number: %d", appendResp))
	if err != nil {
		t.Errorf(err.Error())
	}
	for resp := range c {
		fmt.Println(fmt.Sprintf("Subscribe responded with global sequence number: %d", resp.Gsn))
	}
}
