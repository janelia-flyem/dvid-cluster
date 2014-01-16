DVID-CLUSTER
============

*Status: In development, not ready for use.*

[![Build Status](https://drone.io/github.com/janelia-flyem/dvid-cluster/status.png)](https://drone.io/github.com/janelia-flyem/dvid-cluster/latest)

DVID-cluster is a cluster-oriented system that loads raw data, processes it, and stores results
as key-value pairs suitable for direct import into a [DVID server](http://github.com/janelia-flyem/dvid).DVID-cluster uses [groupcache](https://github.com/golang/groupcache) to keep big data in RAM.

The user starts the manager, which then spawns node processes across the cluster, determines the
hostnames, and broadcasts the information to the nodes.  The nodes then set their peer group.

While this initial prototype is in Go, the output of the processing is a set of files that can
be imported into a remote DVID server.

## Output file format

TBD