package proj

//var ZkAddrs = []string{"54.197.196.191:2181"}
var ZkAddrs = []string{"localhost:2181"}
var Debug = true
// go from your ip addr to location in sync array
var ReplicaAddrs = map[string]string{"localhost:9600": "localhost:9601", "localhost:9601": "localhost:9602", "localhost:9602": "localhost:9600"}
