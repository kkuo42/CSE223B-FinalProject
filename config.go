package proj

import (
	"time"
)

//var ZkAddrs = []string{"54.197.196.191:2181"}
//var ZkAddrs = []string{"localhost:2181"}
//var ZkAddrs = []string{"vm166.sysnet.ucsd.edu:2181"}
var ZkAddrs = []string{"ubuntu@ec2-34-219-144-16.us-west-2.compute.amazonaws.com"}
var Debug = true
//var Debug = false
// go from your ip addr to location in sync array
var ReplicaAddrs = map[string]string{"localhost:9600": "localhost:9601", "localhost:9601": "localhost:9602", "localhost:9602": "localhost:9600"}
var GeoFSMode = uint32(0755)
var SequentialEphemeral = int32(3)
var BalanceTime = 60 * time.Second

// potential artificial latency in milliseconds
var Latency = 100
