package client

import (
	"context"

	rpc "github.com/scalog/scalog-client/scalog/discovery/rpc"
	"google.golang.org/grpc"
)

func discoverServers(port string) []string {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := rpc.NewDiscoveryClient(conn)
	discoveryRequest := &rpc.DiscoverRequest{}

	resp, err := client.DiscoverServers(context.Background(), discoveryRequest)
	if err != nil {
		panic(err)
	}
	return resp.ServerAddresses
}
