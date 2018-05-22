package main

import (
    "flag"
    "log"

    "os"
    "os/signal"

    "proj"
    "github.com/hanwen/go-fuse/fuse/nodefs"
    "github.com/hanwen/go-fuse/fuse/pathfs"
)

func shutdown() {

}

func main() {
	// TODO note even though nodes are ephemeral we may get a response from a dead node,
	// if a node stops responding for a while and comes back online it needs to re register
	// with zookeeper
    flag.Parse()
    if len(flag.Args()) < 1 {
	    log.Fatal("Usage:\n  fs-front <MOUNTPOINT>")
    }
    mountpoint := flag.Arg(0)

    zkaddr := []string{"54.197.196.191:2181"}

    // setup frontend filesystem
    frontend := proj.NewFrontendRemotelyBacked(zkaddr) // remote

    nfs := pathfs.NewPathNodeFs(&frontend, nil)

    // mount
    server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
    if err != nil {
        log.Fatalf("Mount fail: %v\n", err)
    }

    // before serving catch ^C and cleanly bail out
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
	// signal is a ^C, unmount to shutdown cleanly
	<-c
	log.Printf("unmounting %v", mountpoint)
	server.Unmount()
	os.Exit(1)
    }()

    log.Printf("filesystem store serving to directory \"%s\n", mountpoint)
    server.Serve()
}
