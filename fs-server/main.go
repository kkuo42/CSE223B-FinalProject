package main

import (
    "flag"
    "log"

    "proj"

    "net/rpc"
    "net"
    "net/http"
    "io/ioutil"
    "strings"
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
    flag.Parse()
    addr := ":9898"

    if len(flag.Args()) < 1 {
        log.Fatal("Usage:\n  fs-server SHAREPOINT")
    }
    if len(flag.Args()) == 2 {
	addr = flag.Arg(1)
    }

    pubaddr := pubIP()+":"+strings.Split(addr, ":")[1]
    log.Println("public addr", pubaddr)
    // setup loopback filesystem
    zkaddr := "54.197.196.191:2181"
    sharepoint := flag.Arg(0)
    nfs := proj.NewServerFs(sharepoint, addr, pubaddr, zkaddr)

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
