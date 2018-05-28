package proj

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/samuel/go-zookeeper/zk"
	"encoding/json"
	"time"
	"log"
        "fmt"
)

type ServerMeta struct {
    primaryFor map[string]string
    replicaFor map[string]string
}

type ServerFileMeta struct {
    Addr string
    WriteCount int
    ReadCount int
}

// this struct is to maintain all information relative to the keeper
type KeeperMeta struct {
	Attr fuse.Attr
	Primary ServerFileMeta
	Replicas map[string]ServerFileMeta
}

type KeeperClient struct {
	addr string
	client *zk.Conn
        backends []*ClientFs
        serverMeta ServerMeta
}

type KeeperHandler struct {
	// TODO, when servers go down/up/move 
}

func NewKeeperClient(addr string) *KeeperClient {
	return &KeeperClient{addr: addr}
}

func (k *KeeperClient) Connect() error {
	client, _, err := zk.Connect(ZkAddrs, time.Second)
        // TODO if cant connect go find another server?
        if err != nil { return err }
	k.client = client
        return nil
}

func (k *KeeperClient) Init() error {
        if k.client == nil {
            e := k.Connect()
            if e != nil {
		log.Fatalf("error connecting to zkserver\n")
		return e
            }
        }

	// attempt to create alive and data dirs, if it fails itll be caught below
	_, _= k.client.Create("/alive", []byte("alive"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/alivemeta", []byte("alivemeta"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/data", []byte("data"), int32(0), zk.WorldACL(zk.PermAll))

        // if a server is joining (addr != "") then create the node in zk 
        if k.addr != "" {
            _, err := k.client.Create("/alive/"+k.addr, []byte(k.addr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
            empty := map[string]string{}
            sm := ServerMeta{empty, empty}
            d, e := json.Marshal(&sm)
            if e != nil { return e }
            _, err = k.client.Create("/alivemeta/"+k.addr, d, int32(0), zk.WorldACL(zk.PermAll))
            if err != nil {
                    log.Fatalf("error creating node in zkserver")
                    return err
            }
        }

        // after getting all backends set up the clients for each
        backs, e := k.GetBackends()
        if e != nil {
                log.Fatalf("error getting backends", e)
                return e
        }
        k.backends = backs
        go k.Watch()
	return nil
}

func (k *KeeperClient) Watch() {
    for {
        backs, stat, watch, e := k.client.ChildrenW("/alive")
        if e != nil { panic("some real bad happened") }
        // TODO probably should go and look at all nodes to avoid corner cases.. kind of ugly.
        if len(backs) < len(k.backends) {
            e = k.HandleFail()
        } else if len(backs) > len(k.backends) {
            fmt.Println("node join")
            e = k.UpdateBackends()
            if e != nil { panic("some real bad happened") }
        }
        blah := <-watch
        fmt.Println("blah data with stat", blah, stat)
    }
}

func (k *KeeperClient) HandleFail() error {
    fmt.Println("node failure")
    e := k.UpdateBackends()
    if e != nil { panic("something real bad happened") }
    return e
}

func (k *KeeperClient) GetBackends() ([]*ClientFs, error) {
    e := k.UpdateBackends()
    if e != nil { return nil, e }
    return k.backends, nil
}

func (k *KeeperClient) UpdateBackends() (error) {
    // this gets /alive and pings them in order, returning a list of addrs
    // in the order that they respond.
    backends, _, _, e := k.client.ChildrenW("/alive")
    if e != nil {
            log.Fatalf("error getting alive nodes", e)
            return e
    }
    done := make(chan bool)
    backs := []*ClientFs{}

    for _, addr := range backends {
        go func(a string) {
            c := NewClientFs(a)
            // if it is not this server then connect
            if a != k.addr {
                e := c.Connect()
                if e != nil {
                        log.Println("keeper couldnt connect to backend", addr)
                }
            }
            backs = append(backs, c)
            done <- true
        }(addr)
    }

    for i := 0; i < len(backends); i++ {
        <-done
    }

    if len(backs) == 0 {
        fmt.Errorf("No backends online yet\n")
    }
    k.backends = backs
    return nil
}

func (k *KeeperClient) Get(path string) (KeeperMeta, error) {
	data, _, e := k.client.Get("/data/" + path)
	if e != nil {
		// do nothing for now, should crash
		return KeeperMeta{}, e
	}
	var kmeta KeeperMeta
	e = json.Unmarshal(data, &kmeta)
	return kmeta, nil
}

func (k *KeeperClient) Set(path string, data KeeperMeta) error {
	m, e := json.Marshal(&data)
	_, e = k.client.Set("/data/"+path, m, -1)
	if e != nil {
		return e
	}
	return nil
}

func (k *KeeperClient) GetChildren(path string) ([]string, error) {
	// if in the root directory don't add /
	inputstr := "/data"
	if path != "" {
		inputstr += "/"+path
	}

	files, _, e := k.client.Children(inputstr)
	if e != nil {
		return nil, e
	}

	return files, nil
}

func (k *KeeperClient) GetChildrenAttributes(path string) ([]fuse.DirEntry, error) {
	files, e := k.GetChildren(path)

	if e != nil {
		return nil, e
	}
	fileEntries := []fuse.DirEntry{}

	// for each of the files fetched get their metadata (may be unecessary)
	for _, f := range files {
		p := f
		if path != "" {
			p = path + "/" + f
		}
		fm, e := k.Get(p)
		if e != nil {
			log.Println("error here:", e)
			return nil, e
		}
		fileEntries = append(fileEntries, fuse.DirEntry{Name: f, Mode: fm.Attr.Mode})
	}
	return fileEntries, nil
}

func (k *KeeperClient) Create(path string, attr fuse.Attr) error {
        // brand new file, initialize new file metadata
        primary := ServerFileMeta{k.addr, 0, 0}
        // pick a replica on the median
        replica := k.backends[len(k.backends)/2].Addr
        var kmeta KeeperMeta
        if replica != k.addr {
            replicas := map[string]ServerFileMeta{replica: ServerFileMeta{replica, 0, 0}}
            kmeta = KeeperMeta{Primary: primary, Replicas: replicas, Attr: attr}
        } else {
            kmeta = KeeperMeta{Primary: primary, Attr: attr}
        }
	d, e := json.Marshal(&kmeta)
	if e != nil {
		return e
	}

	_, e = k.client.Create("/data/"+path, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
	if e != nil {
		return e
	}
        k.AddServerMeta(path, true)
	return nil
}

func (k *KeeperClient) Remove(path string) error {
	err := k.client.Delete("/data/"+path, -1)

	if err != nil {
		return err
	}
	return nil
}

func (k *KeeperClient) RemoveDir(path string) error {
	// recursively delete all children
	children, e := k.GetChildren(path)
	if len(children) == 0 || e != nil {
		// it is safe to delete this node
		k.Remove(path)
	} else {
		// go through each child and recursively delete on it
		for _, f := range children {
			e := k.RemoveDir(path+"/"+f)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func (k *KeeperClient) Inc(path string, read bool) error {
    kmeta, e := k.Get(path)
    if e != nil { return e }
    // if this server is not the primary go increment in replica
    if k.addr != kmeta.Primary.Addr {
        sfm := kmeta.Replicas[k.addr]
        if read {
            sfm.ReadCount += 1
        } else {
            sfm.WriteCount += 1
        }
        kmeta.Replicas[k.addr] = sfm
    } else {
        if read {
            kmeta.Primary.ReadCount += 1
        } else {
            kmeta.Primary.WriteCount += 1
        }
    }
    e = k.Set(path, kmeta)
    if e != nil { return e }
    return nil
}

/*
func (k *KeeperClient) EditServer(path string, replica bool) error {
}*/

// add a primary or a replica to a server's metadata
func (k *KeeperClient) AddServerMeta(path string, replica bool) error {
    smeta, _, e := k.client.Get("/servermeta/" + path)
    if e != nil { return e }

    var sm ServerMeta
    e = json.Unmarshal(smeta, &sm)
    if e != nil { return e }
    if replica {
        sm.replicaFor[path] = path
    } else {
        sm.primaryFor[path] = path
    }
    smdata, e := json.Marshal(&sm)
    if e != nil { return e }
    _, e = k.client.Set("/alivemeta/"+k.addr, smdata, -1)
    if e != nil { return e }
    return nil
}

func (k *KeeperClient) RemoveServerMeta(path string, replica bool) error {
    smeta, _, e := k.client.Get("/servermeta/" + path)
    if e != nil { return e }

    var sm ServerMeta
    e = json.Unmarshal(smeta, &sm)
    if e != nil { return e }
    if replica {
        delete(sm.replicaFor, path)
    } else {
        delete(sm.primaryFor, path)
    }
    smdata, e := json.Marshal(&sm)
    if e != nil { return e }
    _, e = k.client.Set("/alivemeta/"+k.addr, smdata, -1)
    if e != nil { return e }
    return nil
}
