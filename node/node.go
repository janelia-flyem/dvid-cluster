package node

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"

	_ "github.com/golang/groupcache"

	_ "github.com/janelia-flyem/dvid/dvid"
	_ "github.com/janelia-flyem/dvid/server"

	// Declare the data types this executable will support
	_ "github.com/janelia-flyem/dvid/datatype/keyvalue"
	_ "github.com/janelia-flyem/dvid/datatype/labelmap"
	_ "github.com/janelia-flyem/dvid/datatype/labels64"
	_ "github.com/janelia-flyem/dvid/datatype/multichan16"
	_ "github.com/janelia-flyem/dvid/datatype/tiles"
	_ "github.com/janelia-flyem/dvid/datatype/voxels"
)

type Peers struct {
	Hostnames []string
}

// RPCConnection will export all of its functions for rpc access.
type RPCConnection struct{}

const RPCAddress = ":8001"

var peers []string

func Serve() {
	rpcServer := new(RPCConnection)
	rpc.Register(rpcServer)
	rpc.HandleHTTP()

	address := fmt.Sprintf("%s", RPCAddress)
	fmt.Printf("Listening on %s\n", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening on %s: %s\n", RPCAddress, err.Error())
		os.Exit(1)
	}
	http.Serve(listener, nil)
}

func (c *RPCConnection) SetPeers(arg *Peers, reply *int) error {
	peers = arg.Hostnames
	fmt.Printf("Set peers to %s\n", peers)
	return nil
}

