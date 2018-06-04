package main

import (
    "flag"
    "log"

    "proj"

    "time"
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
    // parse args
    flag.Parse()
    if len(flag.Args()) != 3 {
        log.Fatal("Usage:\n  fs-server <SHAREPOINT> <COORD ADDR> <SERVERFS ADDR>")
    }
    sharepoint := flag.Arg(0)
    coordaddr := flag.Arg(1)
    fsaddr := flag.Arg(2)
    port := strings.Split(coordaddr, ":")[1]

    // setup server coordinator 
    nsc := proj.NewServerCoordinator(sharepoint, coordaddr, fsaddr)
    go func() {
        // wait for this to serve before keeper init
        time.Sleep(time.Second / 2)
        nsc.Init()
    }()

    // setup rpc server
    server := rpc.NewServer()
    e := server.RegisterName("BackendFs", nsc)
    l, e := net.Listen ("tcp",":"+port)
    if e != nil {
        log.Fatal(e)
    }

    // serve
    log.Printf("key-value coordinator on %s", coordaddr)
    e = http.Serve(l, server)
    if e != nil {
        log.Fatal(e)
    }
}
