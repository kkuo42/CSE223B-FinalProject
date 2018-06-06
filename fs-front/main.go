package main

import (
    "proj"
    "flag"
    "log"
)

func main() {
	// TODO note even though nodes are ephemeral we may get a response from a dead node,
	// if a node stops responding for a while and comes back online it needs to re register
	// with zookeeper
    backaddr := ""

    // parse args
    flag.Parse()
    if len(flag.Args()) < 1 {
	    log.Fatal("Usage:\n  fs-front <MOUNTPOINT> <optional: BACKENDIP>")
    }
    if len(flag.Args()) == 2 {
	    backaddr = flag.Arg(1)
    }
    mountpoint := flag.Arg(0)

    fsfront := proj.NewFsfront(backaddr, mountpoint)
    fsfront.Run()
}
