package main

import (
    "proj"
    "flag"
    "log"
    "io/ioutil"
    "net/http"
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
	argc := len(flag.Args())
    if argc != 1 && argc != 3 {
        log.Fatal("Usage:\n  fs-server <SHAREPOINT> <optional COORD ADDR> <optional SERVERFS ADDR>")
    }
	var coordaddr string
	var fsaddr string

	if argc == 1 {
		coordaddr = pubIP() + ":9898"
		fsaddr = pubIP() + ":9899"
	}

	if argc == 3 {
		coordaddr = flag.Arg(1)
		fsaddr = flag.Arg(2)
	}

    sharepoint := flag.Arg(0)

    fsserver := proj.NewFsserver(sharepoint, coordaddr, fsaddr, nil)
    fsserver.Run()
}
