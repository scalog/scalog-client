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

func NewIt() (*It, error) {
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	it := &It{}
	it.client = client
	return it, nil
}

func (it *It) Start() error {
	regex := regexp.MustCompile(" +")
	reader := bufio.NewReader(os.Stdin)
	for {
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		cmdString = strings.TrimSuffix(cmdString, "\n")
		cmdString = strings.Trim(cmdString, " ")
		if cmdString == "" {
			continue
		}
		cmd := regex.Split(cmdString, -1)
		if cmd[0] == "quit" || cmd[0] == "exit" {
			break
		}
		if cmd[0] == "append" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [record]")
				continue
			}
			record := strings.Join(cmd[1:], " ")
			gsn, err := it.client.Append(record)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "Append result: { Gsn: %d, Record: %s }\n", gsn, record)
		} else if cmd[0] == "subscribe" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [gsn]")
				continue
			}
			gsn, err := strconv.ParseInt(cmd[1], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			subscribeChan, err := it.client.Subscribe(int32(gsn))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			go func() {
				committedRecord := <-subscribeChan
				fmt.Fprintf(os.Stderr, "Subscribe result: { Gsn: %d, Record: %s }\n", committedRecord.Gsn, committedRecord.Record)
			}()
		} else if cmd[0] == "trim" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [gsn]")
				continue
			}
			gsn, err := strconv.ParseInt(cmd[1], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			err = it.client.Trim(int32(gsn))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintln(os.Stderr, "Trim result: {}")
		} else {
			fmt.Fprintln(os.Stderr, "Command error: invalid command")
		}
	}
	return nil
}
