package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/janelia-flyem/drmaa"
	"github.com/janelia-flyem/drmaa/gestatus"

	"github.com/janelia-flyem/dvid/dvid"
	"github.com/janelia-flyem/dvid/server"

	"github.com/janelia-flyem/dvid-cluster/node"
)

var (
	// Display usage if true.
	showHelp = flag.Bool("help", false, "")

	// Run in debug mode if true.
	runDebug = flag.Bool("debug", false, "")

	// Address for rpc communication.
	rpcAddress = flag.String("rpc", server.DefaultRPCAddress, "")

	// Address for http communication
	httpAddress = flag.String("http", server.DefaultWebAddress, "")

	email = flag.String("email", "", "")
)

const helpMessage = `
%s manages DVID cluster nodes for processing data.

Usage: %s [options] <num nodes> <command>

      -rpc        =string   Address for RPC communication.
      -http       =string   Address for HTTP communication.
      -email      =string   Email address for job notification
      -debug      (flag)    Run in debug mode.  Verbose.
  -h, -help       (flag)    Show help message

`

func currentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Could not get current directory:", err)
	}
	return currentDir
}

// Waits for running job, determines hostname and sends it down channel.
func waitForRunning(s drmaa.Session, jobId string, hostCh chan string) {
	// Wait for running job
	d, _ := time.ParseDuration("500ms")
	ps, _ := s.JobPs(jobId)
	for ps != drmaa.PsRunning {
		time.Sleep(d)
		ps, _ = s.JobPs(jobId)
	}

	// Get hostname
	jobStatus, err := gestatus.GetJobStatus(&s, jobId)
	if err != nil {
		fmt.Printf("Error in getting hostname for job %s: %s\n", jobId, err.Error())
		hostCh <- ""
		return
	}
	hostname := jobStatus.DestinationHostList()
	hostCh <- strings.Join(hostname, "")
}

func main() {
	flag.BoolVar(showHelp, "h", false, "Show help message")
	flag.Usage = func() {
		fmt.Printf(helpMessage, os.Args[0], os.Args[0])
	}
	flag.Parse()

	if flag.NArg() >= 1 && strings.ToLower(flag.Args()[0]) == "help" {
		*showHelp = true
	}
	if flag.NArg() < 2 {
		*showHelp = true
	}

	if *runDebug {
		dvid.Mode = dvid.Debug
	}
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Get the command
	numNodes, err := strconv.Atoi(flag.Args()[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Illegal # of nodes specified '%s': %s\n", flag.Args()[0], err)
		os.Exit(1)
	}
	if numNodes < 1 {
		fmt.Fprintln(os.Stderr, "Need at least one node specified in arguments!")
		flag.Usage()
		os.Exit(0)
	}
	command := flag.Args()[1:]

	// Create a new DRMAA1 session.
	s, _ := drmaa.MakeSession()
	defer s.Exit()

	// Submit the node processes to the cluster.
	jobIds := make([]string, numNodes)
	hostCh := make(chan string, numNodes)

	jt, _ := s.AllocateJobTemplate()
	jt.SetRemoteCommand("/groups/flyem/proj/builds/cluster2014/bin/dvid-node")
	jt.SetEmail([]string{*email})

	options := fmt.Sprintf("-pe batch %d", 16)
	jt.SetNativeSpecification(options)

	for n := 0; n < numNodes; n++ {
		jt.SetArgs(command)
		jobIds[n], _ = s.RunJob(&jt)
		go waitForRunning(s, jobIds[n], hostCh)
	}

	// Gather all the running hostnames
	hostnames := []string{}
	for n := 0; n < numNodes; n++ {
		hostname := <-hostCh
		if hostname == "" {
			fmt.Println("Could not start node!")
		} else {
			fmt.Printf("Started node on %s (%d/%d)\n", hostname, n+1, numNodes)
			hostnames = append(hostnames, hostname)
		}
	}

	// Wait until servers have spun up, preventing race condition on sending peers.
	time.Sleep(10 * time.Second)

	// Relay set of hostnames to all nodes.
	arg := node.Peers{hostnames}
	for _, hostname := range hostnames {
		address := fmt.Sprintf("%s%s", hostname, node.RPCAddress)
		client, err := rpc.DialHTTP("tcp", address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Did not find node at %s: %s\n", address, err.Error())
		} else {
			var reply int
			err = client.Call("RPCConnection.SetPeers", &arg, &reply)
			if err != nil {
				fmt.Fprintf(os.Stderr, "RPC error to %s: %s\n", address, err.Error())
			} else {
				fmt.Printf("Successfully sent %d peers to %s\n", len(hostnames), address)
			}
		}
	}
}
