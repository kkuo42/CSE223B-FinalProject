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

    "fmt"
    // "time"
    "strconv"
    // "os"
    "os/exec"
    // "log"
    "bytes"
    // "proj"
    "github.com/hanwen/go-fuse/fuse"

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





type tester struct {
    pairs int
    Coord []*ServerCoordinator
}

func NewTester(pairs int) *tester {
    retval := &tester{}
    retval.setupZK()

    os.RemoveAll("data")
    os.Mkdir("data", os.ModePerm)
    
    var coord []*ServerCoordinator
    for j := 0; j<pairs; j++ {
        coord = append(coord, retval.setupPair(j))
    }
    retval.Coord = coord

    return retval
}

func (self *tester) Stop() {
    cmd := exec.Command("zookeeper-3.4.12/bin/zkServer.sh", "stop")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf(out.String())

    for j := 0; j<self.pairs; j++ {
        self.stopPair(j)
    }
}

func (self *tester) setupZK() {
    os.RemoveAll("zkdata")
    os.Remove("zookeeper.out")
    os.Mkdir("zkdata", os.ModePerm)

    cmd := exec.Command("zookeeper-3.4.12/bin/zkServer.sh", "start")
    var out bytes.Buffer
    cmd.Stdout = &out   

    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf(out.String())
}

func (self *tester) setupPair(i int) *ServerCoordinator {

    insert := strconv.Itoa(i)

    sharepoint := "data/from" + insert
    os.Mkdir(sharepoint, os.ModePerm)
    coordaddr := "localhost:950" + insert
    fsaddr := "localhost:960" + insert
    fmt.Println()
    retvalChannel := make(chan *ServerCoordinator)

    fsserver := NewFsserver(sharepoint, coordaddr, fsaddr, retvalChannel)
    go fsserver.Run()
    retval := <- retvalChannel
    time.Sleep(time.Second)

    // backaddr := coordaddr
    // mountpoint := "data/to" + insert
    // os.Mkdir(mountpoint, os.ModePerm)
    // fmt.Println()
 //    go fsfront.Run(backaddr, mountpoint)
 //    time.Sleep(time.Second)
    return retval
}

func (self *tester) stopPair(i int) {
    cmd := exec.Command("fusermount", "-u", "data/to"+strconv.Itoa(i))
    var out bytes.Buffer
    cmd.Stdout = &out   
    cmd.Run()
}

func (self *tester) CreateFile(coord_id int, path string) {
    input := &Create_input{Path: path}
    output := &Create_output{}
    e := self.Coord[coord_id].Create(input, output)
    if e != nil || output.Status != fuse.OK {
        panic(e)
    }
}

func (self *tester) AssertFileExists(path string) {
    fmt.Println(path)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        panic("file does not exist " + path)
    }

}
func (self *tester) AssertFileNotExists(path string) {
    fmt.Println(path)
    if _, err := os.Stat(path); err == nil {
        panic("file does exists " + path)
    }
}
func (self *tester) Assert(something bool) {
    if !something {
        panic("statement is false")       
    }
}




