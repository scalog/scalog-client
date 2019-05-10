# Scalog Client

A Go Client library for the [Scalog](https://github.com/scalog/scalog) Project.

[![Build Status](https://travis-ci.org/scalog/scalog-client.svg?branch=master)](https://travis-ci.org/scalog/scalog-client)

## Setup

The following assumes [Go](https://golang.org/) and [dep](https://github.com/golang/dep) are already installed and configured on your machine.

Navigate to the project root directory.

Rename the file `config.yaml.template` to `config.yaml`.

In `config.yaml`, set the IP and port on which the Scalog discovery service is running.

```
discovery-address:
  ip:   "127.0.0.1" // Set the IP
  port: 8000        // Set the port
```

Run the below command in the root directory to update dependencies.

`dep ensure`

Run the below commands to build the project.

```
cd client
go build
```

## Example Usage

```go
import "github.com/scalog/scalog-client/client"

client := client.NewClient()
resp, err := client.Append("Hello, World!")
if err != nil {
    fmt.Println(err.Error())
}
fmt.Println(fmt.Sprintf("Global sequence number %d assigned to record", resp))
```
