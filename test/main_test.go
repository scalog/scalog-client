package test

import (
	"fmt"
	"testing"

	"github.com/scalog/scalog-client/client"
)

func TestClientAppend(t *testing.T) {
	client := client.NewClient()
	resp := client.Append("Hello, World!")
	fmt.Println(fmt.Sprintf("%d", resp))
}
