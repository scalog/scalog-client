// Package lib provides applications with the ability to create an instance of
// Scalog client, which interacts with the Scalog API.
package lib

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	data "github.com/scalog/scalog/data/messaging"
	discovery "github.com/scalog/scalog/discovery/rpc"
	"google.golang.org/grpc"
)

// CommittedRecord represents a record that has been commited by Scalog.
type CommittedRecord struct {
	// Global sequence number assigned by Scalog
	Gsn int32
	// Data of record
	Record string
}

// ShardPolicy determines which records are appended to which shards.
type ShardPolicy func(servers []*discovery.DataServer, record string) (server *discovery.DataServer)

// Client interacts with the Scalog API.
type Client struct {
	// Unique client identifier
	clientID int32
	// Client sequence number to be assigned to the next record
	nextCsn int32
	// Mutex for accessing nextCsn
	appendMu sync.RWMutex
	// Global sequence number of next CommitedRecord to respond to if subscribed
	nextGsn int32
	// Map from global sequence number to CommittedRecord
	committedRecords map[int32]CommittedRecord
	// Mutex for accessing nextGsn and commitedRecords
	subscribeMu sync.RWMutex
	// Channel to send CommitedRecords if subscribed
	subscribeChan chan CommittedRecord
	// Function that determines which records are appended to which shards
	shardPolicy ShardPolicy
	// Configuration meta-data specified in config.yaml
	config *config
}

// address represents an IP address and port number.
type address struct {
	IP   string `yaml:"ip"`
	Port int32  `yaml:"port"`
}

// config contains the meta-data specified in config.yaml.
type config struct {
	DiscoveryAddress address `yaml:"discovery-address"`
}

// NewClient returns a new instance of Client.
func NewClient() (*Client, error) {
	config, err := parseConfig()
	if err != nil {
		return nil, err
	}
	c := &Client{
		clientID:         assignClientID(),
		nextCsn:          0,
		appendMu:         sync.RWMutex{},
		nextGsn:          -1,
		committedRecords: make(map[int32]CommittedRecord),
		subscribeMu:      sync.RWMutex{},
		subscribeChan:    make(chan CommittedRecord),
		shardPolicy:      defaultShardPolicy,
		config:           config,
	}
	return c, nil
}

// Append appends a record to a shard based on the shard policy, and returns the
// global sequence number assigned by Scalog.
func (c *Client) Append(record string) (int32, error) {
	gsn, _, err := c.AppendToShard(record)
	if err != nil {
		return -1, err
	}
	return gsn, nil
}

// AppendToShard appends a record to a shard based on the shard policy, and
// returns the global sequence number assigned by Scalog and the shard's
// identifier.
func (c *Client) AppendToShard(record string) (int32, int32, error) {
	servers, err := c.discoverServers()
	if err != nil {
		return -1, -1, err
	}
	server := c.shardPolicy(servers, record)
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// conn, err := grpc.Dial(getAddressOfServer(server), opts...)
	// TODO: temporary fix due to discovery service returning server's cluster IP
	// TODO: don't dial for every operation. Save the connection and reuse it
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.config.DiscoveryAddress.IP, server.Port), opts...)
	if err != nil {
		return -1, -1, err
	}
	defer conn.Close()
	dataClient := data.NewDataClient(conn)
	c.appendMu.Lock()
	req := &data.AppendRequest{
		Cid:    c.clientID,
		Csn:    c.nextCsn,
		Record: record,
	}
	c.nextCsn++
	c.appendMu.Unlock()
	resp, err := dataClient.Append(context.Background(), req)
	if err != nil {
		return -1, -1, err
	}
	return resp.Gsn, server.ShardID, nil
}

// Subscribe subscribes to CommitedRecords starting from a global sequence
// number, and returns a channel on which to read from.
func (c *Client) Subscribe(gsn int32) (chan CommittedRecord, error) {
	servers, err := c.discoverServers()
	if err != nil {
		return nil, err
	}
	c.subscribeMu.Lock()
	c.nextGsn = gsn
	c.subscribeMu.Unlock()
	for _, server := range servers {
		// go c.subscribeToServer(server, gsn)
		// TODO: temporary fix due to discovery service returning server's cluster IP
		go c.subscribeToServer(
			&discovery.DataServer{
				Ip:   c.config.DiscoveryAddress.IP,
				Port: server.Port,
			},
			gsn,
		)
	}
	return c.subscribeChan, nil
}

// ReadRecord reads a record with a global sequence number from a shard.
func (c *Client) ReadRecord(gsn int32, shardID int32) (string, error) {
	servers, err := c.discoverServers()
	if err != nil {
		return "", err
	}
	for _, server := range servers {
		if server.ShardID == shardID {
			record, err := c.readFromServer(server, gsn)
			if err != nil {
				return "", err
			}
			return record, nil
		}
	}
	return "", fmt.Errorf("Attempted to read record from non-existant shard %d", shardID)
}

// Trim deletes records before a global sequence number from the data servers.
func (c *Client) Trim(gsn int32) error {
	servers, err := c.discoverServers()
	if err != nil {
		return err
	}
	for _, server := range servers {
		// go c.trim(server, gsn)
		// TODO: temporary fix due to discovery service returning server's cluster IP
		go c.trimFromServer(
			&discovery.DataServer{
				Ip:   c.config.DiscoveryAddress.IP,
				Port: server.Port,
			},
			gsn,
		)
	}
	return nil
}

// SetShardPolicy sets the policy for determining which records are appended to
// which shards.
func (c *Client) SetShardPolicy(shardPolicy ShardPolicy) {
	c.shardPolicy = shardPolicy
}

// assignClientID returns a randomly generated 31-bit integer as int32.
func assignClientID() int32 {
	seed := rand.NewSource(time.Now().UnixNano())
	return rand.New(seed).Int31()
}

// defaultShardPolicy returns a random server.
func defaultShardPolicy(servers []*discovery.DataServer, record string) *discovery.DataServer {
	seed := rand.NewSource(time.Now().UnixNano())
	return servers[rand.New(seed).Intn(len(servers))]
}

// parseConfig initializes and returns an instance of config with the meta-data
// specified in config.yaml.
func parseConfig() (*config, error) {
	file, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}
	var config config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// subscribeToServer subscribes to a data server and sends CommittedRecords in
// order to the subscribeChan
func (c *Client) subscribeToServer(server *discovery.DataServer, gsn int32) error {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// conn, err := grpc.Dial(getAddressOfServer(server), opts...)
	// TODO: temporary fix due to discovery service returning server's cluster IP
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.config.DiscoveryAddress.IP, server.Port), opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	dataClient := data.NewDataClient(conn)
	req := &data.SubscribeRequest{SubscriptionGsn: gsn}
	stream, err := dataClient.Subscribe(context.Background(), req)
	if err != nil {
		return err
	}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		c.subscribeMu.Lock()
		c.committedRecords[in.Gsn] = CommittedRecord{
			Gsn:    in.Gsn,
			Record: in.Record,
		}
		if in.Gsn == c.nextGsn {
			c.respond()
		}
		c.subscribeMu.Unlock()
	}
}

// respond sends CommitedRecords to the subscribeChan in order of global sequence
// number starting from nextGsn.
func (c *Client) respond() {
	for commitedRecord, in := c.committedRecords[c.nextGsn]; ; c.nextGsn++ {
		if !in {
			break
		}
		c.subscribeChan <- commitedRecord
		delete(c.committedRecords, c.nextGsn)
	}
}

// trimFromServer deletes records before a global sequence number from a data
// server.
func (c *Client) trimFromServer(server *discovery.DataServer, gsn int32) error {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// conn, err := grpc.Dial(getAddressOfServer(server), opts...)
	// TODO: temporary fix due to discovery service returning server's cluster IP
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.config.DiscoveryAddress.IP, server.Port), opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	dataClient := data.NewDataClient(conn)
	req := &data.TrimRequest{Gsn: gsn}
	_, err = dataClient.Trim(context.Background(), req)
	return err
}

// readFromServer reads a record with a global sequence number from a server.
func (c *Client) readFromServer(server *discovery.DataServer, gsn int32) (string, error) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(c.config.DiscoveryAddress.stats(), opts...)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	dataClient := data.NewDataClient(conn)
	req := &data.ReadRequest{Gsn: gsn}
	resp, err := dataClient.Read(context.Background(), req)
	if err != nil {
		return "", err
	}
	return resp.Record, nil
}

// discoverServers queries the discovery service and returns the live data
// servers.
func (c *Client) discoverServers() ([]*discovery.DataServer, error) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(c.config.DiscoveryAddress.stats(), opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	discoveryClient := discovery.NewDiscoveryClient(conn)
	req := &discovery.DiscoverRequest{}
	resp, err := discoveryClient.DiscoverServers(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return resp.Servers, nil
}

// getAddressOfServer returns the address of a server as a string.
func getAddressOfServer(server *discovery.DataServer) string {
	return fmt.Sprintf("%s:%d", server.Ip, server.Port)
}

// stats returns an address as a string
func (a address) stats() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}
