package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/golang/groupcache"

	"github.com/janelia-flyem/dvid/datastore"
	"github.com/janelia-flyem/dvid/dvid"
	"github.com/janelia-flyem/dvid/server"

	// Declare the data types this DVID executable will support
	_ "github.com/janelia-flyem/dvid/datatype/keyvalue"
	_ "github.com/janelia-flyem/dvid/datatype/labelmap"
	_ "github.com/janelia-flyem/dvid/datatype/labels64"
	_ "github.com/janelia-flyem/dvid/datatype/multichan16"
	_ "github.com/janelia-flyem/dvid/datatype/tiles"
	_ "github.com/janelia-flyem/dvid/datatype/voxels"
)

var (
	// Display usage if true.
	showHelp = flag.Bool("help", false, "")

	// Run in debug mode if true.
	runDebug = flag.Bool("debug", false, "")

	// Run in benchmark mode if true.
	runBenchmark = flag.Bool("benchmark", false, "")

	// Profile CPU usage using standard gotest system.
	cpuprofile = flag.String("cpuprofile", "", "")

	// Profile memory usage using standard gotest system.
	memprofile = flag.String("memprofile", "", "")

	// Path to web client directory.  Leave unset for default pages.
	clientDir = flag.String("webclient", "", "")

	// Address for rpc communication.
	rpcAddress = flag.String("rpc", server.DefaultRPCAddress, "")

	// Address for http communication
	httpAddress = flag.String("http", server.DefaultWebAddress, "")

	// Number of logical CPUs to use for DVID.
	useCPU = flag.Int("numcpu", 0, "")

	// Accept and send stdin to server for use in commands if true.
	useStdin = flag.Bool("stdin", false, "")
)

const helpMessage = `
%s loads, processes, and emits key-value pairs corresponding to some subspace.

Usage: %s [options]  <uuid> <command>

      -rpc        =string   Address for RPC communication.
      -http       =string   Address for HTTP communication.
      -cpuprofile =string   Write CPU profile to this file.
      -memprofile =string   Write memory profile to this file on ctrl-C.
      -numcpu     =number   Number of logical CPUs to use for this node.
      -stdin      (flag)    Accept and send stdin to server for use in commands.
      -debug      (flag)    Run in debug mode.  Verbose.
      -benchmark  (flag)    Run in benchmarking mode. 
  -h, -help       (flag)    Show help message

  For profiling, please refer to this excellent article:
  http://blog.golang.org/2011/06/profiling-go-programs.html

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
	if *runBenchmark {
		dvid.Mode = dvid.Benchmark
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Determine numer of logical CPUs on local machine and unless overridden, use
	// all of them.
	numCPU := runtime.NumCPU()
	if *useCPU != 0 {
		dvid.NumCPU = *useCPU
	} else if flag.NArg() >= 1 && flag.Args()[0] == "serve" {
		dvid.NumCPU = numCPU
	} else {
		dvid.NumCPU = 1
	}
	runtime.GOMAXPROCS(dvid.NumCPU)

	// Capture ctrl+c and other interrupts.  Then handle graceful shutdown.
	stopSig := make(chan os.Signal)
	go func() {
		for sig := range stopSig {
			log.Printf("Stop signal captured: %q.  Shutting down...\n", sig)
			if *memprofile != "" {
				log.Printf("Storing memory profiling to %s...\n", *memprofile)
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Fatal(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
			}
			if *cpuprofile != "" {
				log.Printf("Stopping CPU profiling to %s...\n", *cpuprofile)
				pprof.StopCPUProfile()
			}
			os.Exit(0)
		}
	}()
	signal.Notify(stopSig, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Wait until we discover our peer group.

	// Process a subset of the data.
}
