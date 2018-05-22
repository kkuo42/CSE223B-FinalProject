package main

import (
    "flag"
    "log"

    "proj"

    "net/rpc"
    "net"
    "net/http"
)

func main() {
    flag.Parse()
    if len(flag.Args()) < 1 {
        log.Fatal("Usage:\n  fs-server SHAREPOINT")
    }

    // setup loopback filesystem
    addr := "localhost:9898"
	zkaddr := "localhost:2181"
    nfs := proj.NewServerFs(flag.Arg(0), addr, zkaddr)


    // setup rpc server
    server := rpc.NewServer()
    e := server.RegisterName("BackendFs", &nfs)
    l, e := net.Listen ("tcp", addr)
    if e != nil {
        log.Fatal(e)
    }

    // serve
    log.Printf("key-value store serving directory \"%s\" on %s", flag.Arg(0), addr)
    e = http.Serve(l, server)
    if e != nil {
        log.Fatal(e)
    }
}
