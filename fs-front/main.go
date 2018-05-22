package main

import (
    "flag"
    "log"

    "proj"
    "github.com/hanwen/go-fuse/fuse/nodefs"
    "github.com/hanwen/go-fuse/fuse/pathfs"
)

func main() {
    flag.Parse()
    if len(flag.Args()) < 1 {
        log.Fatal("Usage:\n  fs-front MOUNTPOINT")
    }

    // setup frontend filesystem
    addr := "localhost:9898"
    frontend := proj.NewFrontendRemotelyBacked(addr) // remote
    // dir := "from"
    // frontend := proj.NewFrontendLocalyBacked(dir)    // local
    // frontend := pathfs.NewLoopbackFileSystem(dir)    // local provided loopback

    nfs := pathfs.NewPathNodeFs(&frontend, nil)

    // mount
    server, _, err := nodefs.MountRoot(flag.Arg(0), nfs.Root(), nil)
    if err != nil {
        log.Fatalf("Mount fail: %v\n", err)
    }
    
    log.Printf("filesystem store serving to directory \"%s\"", flag.Arg(0))
    server.Serve()
}
