// Package client provides clients with the ability to interact with the Scalog API.
package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	data "github.com/scalog/scalog/data/messaging"
	discovery "github.com/scalog/scalog/discovery/rpc"
	set64 "github.com/scalog/scalog/pkg/set64"
	"google.golang.org/grpc"
)

/*
SubscribeResponse is a struct that represents a record [Record] that has been ordered
by Scalog and assigned the global sequence number [gsn].
*/
type SubscribeResponse struct {
	Gsn    int32
	Record string
}

/*
Client is a struct that carries the meta-data necessary for a client to interact
with the Scalog API. Clients should initialize a new instance of this struct with
the newClient() function.
*/
type Client struct {
	// Integer serving as unique client id
	cid int32
	// Client-generated record sequence number
	csn int32
	// Mutex for interacting with [csn]
	cmu sync.RWMutex
	// Next global sequence number to respond to client
	nextGsn int32
	// Set of global sequence numbers received from data servers
	ss *set64.Set64
	// Map from global sequence number to [SubscribeResponse]
	sm map[int32]SubscribeResponse
	// Mutex for interacting with [nextGsn], [ss], and [sm]
	smu sync.RWMutex
	// Channel to send subscribe responses to the client
	sc chan SubscribeResponse
}

/*
NewClient initializes and returns a new instance of [Client] with a unique client id.
*/
func NewClient() *Client {
	return &Client{
		cid:     assignClientID(),
		csn:     0,
		cmu:     sync.RWMutex{},
		nextGsn: -1,
		ss:      set64.NewSet64(),
		sm:      make(map[int32]SubscribeResponse),
		smu:     sync.RWMutex{},
		sc:      make(chan SubscribeResponse),
	}
}

/*
Append appends a record [r], and returns the global sequence number assigned to it.
Note: this is a blocking function!
*/
func (c *Client) Append(r string) (int32, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	addresses := discoverServers("130.127.133.24:32403") // TODO: move address to config file
	address := applyAppendPlacementPolicy(addresses)
	conn, err := grpc.Dial(addressToString(address), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := data.NewDataClient(conn)
	c.cmu.Lock()
	appendRequest := &data.AppendRequest{
		Cid:    c.cid,
		Csn:    c.csn,
		Record: r,
	}
	c.csn = c.csn + 1
	c.cmu.Unlock()

	resp, err := dataClient.Append(context.Background(), appendRequest)
	if err != nil {
		return -1, err
	}
	return resp.Gsn, nil
}

/*
Subscribe subscribes to records starting from global sequence number [gsn].
*/
func (c *Client) Subscribe(gsn int32) chan SubscribeResponse {
	c.cmu.Lock()
	c.nextGsn = gsn
	c.cmu.Unlock()
	addresses := discoverServers("130.127.133.24:32403") // TODO: move address to config file
	for _, address := range addresses {
		go c.subscribe(address, gsn)
	}
	return c.sc
}

/*
Trim deletes records before global sequence number [gsn].
*/
func (c *Client) Trim(gsn int32) {
	addresses := discoverServers("130.127.133.24:32403") // TODO: move address to config file
	for _, address := range addresses {
		go c.trim(address, gsn)
	}
}

/*
Returns a randomly generated 31-bit integer as int32.
Note: uniqueness is not guaranteed!
*/
func assignClientID() int32 {
	return rand.Int31()
}

/*
Queries the discovery service and returns a slice of the addresses of all active
data servers.
*/
func discoverServers(address string) []*discovery.DataServerAddress {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	discoveryClient := discovery.NewDiscoveryClient(conn)
	discoveryRequest := &discovery.DiscoverRequest{}

	resp, err := discoveryClient.DiscoverServers(context.Background(), discoveryRequest)
	if err != nil {
		panic(err)
	}
	return resp.Servers
}

/*
Given a slice of addresses of all active data servers, selects and returns a single
address based on the pre-defined data placement policy.
*/
func applyAppendPlacementPolicy(addresses []*discovery.DataServerAddress) *discovery.DataServerAddress {
	if len(addresses) == 0 {
		panic("Failed to append: no active data servers discovered!")
	}
	return addresses[rand.Intn(len(addresses))] // TODO: select address based on data placement policy
}

/*
Creates a stream to the data server on address [address], listens for the
data server to report ordered records with gsn > [gsn], and if possible,
responds to the client in order.
*/
func (c *Client) subscribe(address *discovery.DataServerAddress, gsn int32) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addressToString(address), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := data.NewDataClient(conn)
	subscribeRequest := &data.SubscribeRequest{
		SubscriptionGsn: gsn,
	}

	stream, err := dataClient.Subscribe(context.Background(), subscribeRequest)
	if err != nil {
		panic(err)
	}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}

		c.smu.Lock()
		c.ss.Add(int64(in.Gsn)) // TODO: remove type casting once gsn are int64
		c.sm[in.Gsn] = SubscribeResponse{
			Gsn:    in.Gsn,
			Record: in.Record,
		}
		c.smu.Unlock()
		c.respond()
	}
}

/*
Responds to client with [SubscribeResponse] in order of global sequence number
if possible.
*/
func (c *Client) respond() {
	c.cmu.Lock()
	for c.ss.Contains(int64(c.nextGsn)) { // TODO: remove type casting once gsn are int64
		c.sc <- c.sm[c.nextGsn]
		c.ss.Remove(int64(c.nextGsn)) // TODO: remove type casting once gsn are int64
		delete(c.sm, c.nextGsn)
		c.nextGsn = c.nextGsn + 1
	}
	c.cmu.Unlock()
}

/*
Creates a stream to the data server on address [address] and requests that all
records with gsn < [gsn] be deleted.
*/
func (c *Client) trim(address *discovery.DataServerAddress, gsn int32) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addressToString(address), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := data.NewDataClient(conn)
	r := &data.TrimRequest{
		Gsn: gsn,
	}

	dataClient.Trim(context.Background(), r)
	// No response expected
}

/*
Returns DataServerAddress [address] as a string.
*/
func addressToString(address *discovery.DataServerAddress) string {
	return fmt.Sprintf("%s:%d", address.Ip, address.Port)
}
