package main

import (
    "flag"
    "log"

    "proj"

    "net/rpc"
    "net"
    "net/http"
    "io/ioutil"
)

func pubIP() string {
	// query api to get public ip
	url := "https://api.ipify.org?format=text"
	r, e := http.Get(url)
	if e != nil {
		log.Fatalf("error getting public ip")
		panic(e)
	}
	defer r.Body.Close()
	ip, e := ioutil.ReadAll(r.Body)
	if e != nil {
		log.Fatalf("error getting public ip")
		panic(e)
	}
	return string(ip)
}

func main() {
    // parse args
    flag.Parse()
    if len(flag.Args()) != 3 {
        log.Fatal("Usage:\n  fs-server <SHAREPOINT> <SEVERIP> <ZOOKEEPERIP>")
    }
    sharepoint := flag.Arg(0)
    addr := flag.Arg(1)
    zkaddr := flag.Arg(2)

    // setup loopback filesystem
    nfs := proj.NewServerFs(sharepoint, addr, zkaddr)

    // setup rpc server
    server := rpc.NewServer()
    e := server.RegisterName("BackendFs", &nfs)
    l, e := net.Listen ("tcp", addr)
    if e != nil {
        log.Fatal(e)
    }

    // serve
    log.Printf("key-value store serving directory \"%s\" on %s", sharepoint, addr)
    e = http.Serve(l, server)
    if e != nil {
        log.Fatal(e)
    }
}
