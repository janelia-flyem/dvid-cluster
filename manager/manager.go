package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	// "time"

	"github.com/dgruber/drmaa"
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
)

const helpMessage = `
%s manages DVID cluster nodes for processing data.

Usage: %s [options] <command>

      -rpc        =string   Address for RPC communication.
      -http       =string   Address for HTTP communication.
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

func main() {
	flag.BoolVar(showHelp, "h", false, "Show help message")
	flag.Usage = fmt.Printf(helpMessage, os.Args[0], os.Args[0])
	flag.Parse()

	if flag.NArg() >= 1 && strings.ToLower(flag.Args()[0]) == "help" {
		*showHelp = true
	}

	if *runDebug {
		dvid.Mode = dvid.Debug
	}
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Spawn nodes.

	// Determine the node hostnames.

	// Relay information to all nodes.
}
