package store

import (
    "net/rpc"
    "fmt"
)

type StorageClient struct {
    Addr string
    client *rpc.Client
}

func NewStorageClient(addr string) *StorageClient {
    return &StorageClient{Addr: addr}
}

func (self *StorageClient) Init() error {
    return self.connect()
}

func (self *StorageClient) connect() error {
    fmt.Println("connecting")
    if (self.client == nil) {
        fmt.Println("creating client")
        client, e := rpc.DialHTTP("tcp", self.Addr)
        if e != nil {
            fmt.Println("error dialing", e)
            return e
        }
        fmt.Println("client created")
        self.client = client
    }

    return nil
}

func (self *StorageClient) Stat(fn string) (*StatRes, error) {
    self.connect()
    var res StatRes
    e := self.client.Call("Storage.Stat", fn, &res)
    if e != nil {
        return nil, e
    }
    return &res, nil
}

func (self *StorageClient) Read(fn string) (*ReadRes, error) {
    self.connect()
    var res ReadRes
    e := self.client.Call("Storage.Read", fn, &res)
    if e != nil {
        return nil, e
    }
    return &res, nil
}
