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
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("Global sequence number: %d", resp))
}
