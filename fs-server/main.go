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
    addr := "localhost:9898"

    if len(flag.Args()) < 1 {
        log.Fatal("Usage:\n  fs-server SHAREPOINT")
    }
    if len(flag.Args()) == 2 {
	addr = flag.Arg(1)
    }

    // setup loopback filesystem
    sharepoint := flag.Arg(0)
    nfs := proj.NewServerFs(sharepoint)

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
