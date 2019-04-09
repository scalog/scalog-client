/* Package provides clients with the ability to interact with the Scalog API. */
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
    cid         int32
    // Client-generated record sequence number
    csn         int32
    // Mutex for interacting with [csn]
    mu          sync.RWMutex
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
        cid:    assignClientId(),
        csn:    0,
        mu:     sync.RWMutex{},
    }
}

/*
 * Appends a record [r], and returns the global sequence number assigned to it.
 */
func (c *Client) Append(r string) (int32, error) {
    var opts []grpc.DialOption
    opts = append(opts, grpc.WithInsecure())
    ports := discoverServers("130.127.133.24:32403") // TODO: move port to config file
    port := appendPlacementPolicy(ports)
    conn, err := grpc.Dial(fmt.Sprintf("130.127.133.24:%d", port), opts...)
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    dataClient := data.NewDataClient(conn)
    c.mu.Lock()
    appendRequest := &data.AppendRequest {
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
    ports := discoverServers("130.127.133.24:32403") // TODO: move port to config file
    for _, port := range ports {
        go c.subscribe(port, gsn, callback)
    }
}

/*
 * Deletes records before global sequence number [gsn].
 */
func (c *Client) Trim(gsn int32) {
    // TODO
}

/*
 * Queries the discovery service and returns the unique client id assigned.
 */
func assignClientId() int32 {
    // TODO: query discovery service to obtain client id in order to ensure that
    // client ids are unique
    return 0
}

/*
 * Queries the discovery service and returns the ports of the active data
 * servers as a slice.
 */
func discoverServers(address string) []int32 {
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
    return resp.ServerAddresses
}

/*
 * Given a slice of ports of the active data servers, selects and returns a
 * single port based on the data placement policy.
 */
func appendPlacementPolicy(ports []int32) int32 {
    if len(ports) == 0 {
        panic("Failed to append: no active data servers discovered!")
    }
    return ports[rand.Intn(len(ports))] // TODO: select port based on data placement policy
}

/*
 * Creates a stream to the data server on port [port], listens for the data
 * data server to report ordered records with gsn > [gsn], and passes those
 * arguments to the [callback] function.
 */
func (c *Client) subscribe(port int32, gsn int32, callback Callback) {
    var opts []grpc.DialOption
    opts = append(opts, grpc.WithInsecure())
    conn, err := grpc.Dial(fmt.Sprintf("130.127.133.24:%d", port), opts...)
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
