package main

import (
    "proj"
    "flag"
    "log"

)

func main() {
    // parse args
    flag.Parse()
    if len(flag.Args()) != 3 {
        log.Fatal("Usage:\n  fs-server <SHAREPOINT> <COORD ADDR> <SERVERFS ADDR>")
    }
    sharepoint := flag.Arg(0)
    coordaddr := flag.Arg(1)
    fsaddr := flag.Arg(2)
    
    fsserver := proj.NewFsserver(sharepoint, coordaddr, fsaddr, nil)
    fsserver.Run()
}
