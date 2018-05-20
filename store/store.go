package store

import (
    "os"
    "path/filepath"
    "fmt"
    "time"
    "net"
    "net/rpc"
    "net/http"
)

type StorageServer struct {
    Addr string
    Root string
}

type Storage struct {
    rootdir string
}

type StatRes struct {
    Size int64
    Modified time.Time
    Dir bool
}

type ReadRes struct {
    Size int
    Data []byte
}

func (self *Storage) Stat(req string, res *StatRes) error {
    fmt.Println("stat called", req, self.rootdir)
    path := filepath.Join(self.rootdir, req)
    if s, e := os.Stat(path); os.IsNotExist(e) {
        fmt.Printf("%s does not exist\n", path)
        return e
    } else {
        // the file exists and we want to set the correct information
        res.Modified = s.ModTime()
        if s.IsDir() {
            res.Dir = true
        } else {
            res.Dir = false
            res.Size = s.Size()
        }
    }
    return nil
}

func (self *Storage) Read(req string, res *ReadRes) error {
    var stat StatRes
    e := self.Stat(req, &stat)
    if e != nil {
        return e
    }


    // TODO may need to lock files on open?
    path := filepath.Join(self.rootdir, req)
    file, e := os.Open(path)
    if e != nil {
        return e
    }
    defer file.Close()

    res.Data = make([]byte, stat.Size)
    n, e := file.Read(res.Data)
    if e != nil {
        return e
    }
    res.Size = n
    res.Data = res.Data[:n]
    return nil
}

func ServeStorage(addr, root string) error {
    /*
    server := rpc.NewServer()
    ss := NewStorageServer(addr, root)
    e := server.RegisterName("Storage", ss)
    l, e := net.Listen("tcp", addr)
    if e != nil {
        return e
    }
    return http.Serve(l, server)
    */
    store := new(Storage)
    store.rootdir = root
    rpc.Register(store)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", addr)
    if e != nil {
        return e
    }
    fmt.Printf("listening on %s", addr)
    return http.Serve(l, nil)
}
