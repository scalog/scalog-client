/* Package client provides clients with the ability to interact with the Scalog API. */
package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	data "github.com/scalog/scalog/data/messaging"
	discovery "github.com/scalog/scalog/discovery/rpc"
	"google.golang.org/grpc"
)

/*
 * Struct that carries the meta-data necessary for a client to interact with the
 * Scalog API. Clients should initialize a new instance of this struct with the
 * newClient() function.
 */
type Client struct {
	// Integer serving as unique client id
	cid int32
	// Client-generated record sequence number
	csn int32
	// Mutex for interacting with [csn]
	mu sync.RWMutex
}

/*
 * Function that is called when a data server reports an ordered record to the
 * client after the client has subscribed.
 */
type Callback func(int32, string, error)

/*
 * Initializes and returns a new instance of [Client] with a unique client id.
 */
func NewClient() *Client {
	return &Client{
		cid: assignClientId(),
		csn: 0,
		mu:  sync.RWMutex{},
	}
}

/*
 * Appends a record [r], and returns the global sequence number assigned to it.
 */
func (c *Client) Append(r string) (int32, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	addresses := discoverServers("130.127.133.24:32403") // TODO: move port to config file
	address := appendPlacementPolicy(addresses)
	conn, err := grpc.Dial(addressToString(address), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := data.NewDataClient(conn)
	c.mu.Lock()
	appendRequest := &data.AppendRequest{
		Cid:    c.cid,
		Csn:    c.csn,
		Record: r,
	}
	c.csn = c.csn + 1
	c.mu.Unlock()
	resp, err := dataClient.Append(context.Background(), appendRequest)
	if err != nil {
		return -1, err
	}
	return resp.Gsn, nil
}

/*
 * Subscribes to records starting from global sequence number [gsn].
 */
func (c *Client) Subscribe(gsn int32, callback Callback) {
	addresses := discoverServers("130.127.133.24:32403") // TODO: move port to config file
	for _, address := range addresses {
		go c.subscribe(address, gsn, callback)
	}
}

/*
 * Deletes records before global sequence number [gsn].
 */
func (c *Client) Trim(gsn int32) {
	addresses := discoverServers("130.127.133.24:32403") // TODO: move port to config file
	for _, address := range addresses {
		go c.trim(address, gsn)
	}
}

/*
 * Queries the discovery service and returns the unique client id assigned.
 */
func assignClientId() int32 {
	// TODO: query discovery service to obtain client id in order to ensure that
	// client ids are unique
	return rand.Int31()
}

/*
 * Queries the discovery service and returns the ports of the active data
 * servers as a slice.
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
 * Given a slice of addresses of the active data servers, selects and returns a
 * single address based on the data placement policy.
 */
func appendPlacementPolicy(addresses []*discovery.DataServerAddress) *discovery.DataServerAddress {
	if len(addresses) == 0 {
		panic("Failed to append: no active data servers discovered!")
	}
	return addresses[rand.Intn(len(addresses))] // TODO: select port based on data placement policy
}

/*
 * Creates a stream to the data server on port [port], listens for the data
 * data server to report ordered records with gsn > [gsn], and passes those
 * arguments to the [callback] function.
 */
func (c *Client) subscribe(address *discovery.DataServerAddress, gsn int32, callback Callback) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addressToString(address), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := data.NewDataClient(conn)
	r := &data.SubscribeRequest{
		SubscriptionGsn: gsn,
	}
	stream, err := dataClient.Subscribe(context.Background(), r)
	if err != nil {
		panic(err)
	}

	for true {
		in, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			callback(-1, "", err)
		}
		callback(in.Gsn, in.Record, nil) // TODO: filter duplicate gsn
	}
}

/*
 * Creates a stream to the data server on port [port] and requests that
 * all records with gsn < [gsn] be deleted.
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

func addressToString(address *discovery.DataServerAddress) string {
	return fmt.Sprintf("%s:%d", address.Ip, address.Port)
}
