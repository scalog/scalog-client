package it

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	regex := regexp.MustCompile(" +")

	for {
		reader := bufio.NewReader(os.Stdin)
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		cmdString = strings.TrimSuffix(cmdString, "\n")
		cmdString = strings.Trim(cmdString, " ")
		cmd := regex.Split(cmdString, -1)
		if len(cmd) != 2 {
			fmt.Fprintln(os.Stderr, "Command error")
			continue
		}
		if cmd[0] == "append" {
			gsn, err := it.client.Append(cmd[2])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Println(gsn)
		} else if cmd[0] == "trim" {
			gsn, err := strconv.ParseInt(cmd[2], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			it.client.Trim(int32(gsn))
		}
	}
	return nil
}
