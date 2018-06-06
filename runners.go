package proj

import (
    "log"

    "os"
    "os/signal"

    "github.com/hanwen/go-fuse/fuse/nodefs"
    "github.com/hanwen/go-fuse/fuse/pathfs"


    "time"
    "net/rpc"
    "net"
    "net/http"
    "io/ioutil"
    "strings"
)

type fsfront struct {
    backaddr string
    mountpoint string
}

func NewFsfront(backaddr, mountpoint string) *fsfront {
    return &fsfront{backaddr, mountpoint}
}

func (self *fsfront) Run() {
    log.Printf("fs-front %v %v \n", self.backaddr, self.mountpoint)

    // setup frontend filesystem
    frontend := NewFrontendRemotelyBacked(self.backaddr) // remote
    frontend.Init()
    nfs := pathfs.NewPathNodeFs(&frontend, nil)

    // mount
    server, _, err := nodefs.MountRoot(self.mountpoint, nfs.Root(), nil)
    if err != nil {
        log.Fatalf("Mount fail: %v\n", err)
    }

    // before serving catch ^C and cleanly bail out
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
    	// signal is a ^C, unmount to shutdown cleanly
    	<-c
    	log.Printf("unmounting %v", self.mountpoint)
    	server.Unmount()
    	os.Exit(1)
    }()

    log.Printf("filesystem store serving to directory \"%v\"\n", self.mountpoint)
    server.Serve()
}


type fsserver struct {
    sharepoint string
    coordaddr string
    fsaddr string
    coord chan *ServerCoordinator
}

func NewFsserver(sharepoint, coordaddr, fsaddr string, coord chan *ServerCoordinator) *fsserver {
    return &fsserver{sharepoint, coordaddr, fsaddr, coord}
}


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

func (self *fsserver) Run() {
    log.Printf("fs-server %v %v %v\n", self.sharepoint, self.coordaddr, self.fsaddr)
    // parse args
    port := strings.Split(self.coordaddr, ":")[1]

    // setup server coordinator 
    nsc := NewServerCoordinator(self.sharepoint, self.coordaddr, self.fsaddr)
    if self.coord != nil{
        self.coord <- nsc
    }
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
    log.Printf("key-value coordinator on %s", self.coordaddr)
    e = http.Serve(l, server)
    if e != nil {
        log.Fatal(e)
    }
}