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

func main() {
    flag.Parse()
    addr := "localhost:9898"
    if len(flag.Args()) < 1 {
	    log.Fatal("Usage:\n  fs-front <MOUNTPOINT> <Addr: Optional>")
    }
    if len(flag.Args()) == 2 {
	    addr = flag.Arg(1)
    }
    mountpoint := flag.Arg(0)

    // setup frontend filesystem
    frontend := proj.NewFrontendRemotelyBacked(addr) // remote
    // dir := "from"
    // frontend := proj.NewFrontendLocalyBacked(dir)    // local
    // frontend := pathfs.NewLoopbackFileSystem(dir)    // local provided loopback

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
    }()

    log.Printf("filesystem store serving to directory \"%s\" on %v", mountpoint, addr)
    server.Serve()
}
