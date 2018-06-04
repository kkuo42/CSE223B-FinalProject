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
    PrimaryFor map[string]string
    ReplicaFor map[string]string
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
	coordaddr string
	fsaddr string
	client *zk.Conn
	serverfs []*ClientFs
	servercoords []*ClientFs
}

func NewKeeperClient(coordaddr, fsaddr string) *KeeperClient {
	return &KeeperClient{coordaddr: coordaddr, fsaddr: fsaddr}
}

func (k *KeeperClient) Connect() error {
	client, _, err := zk.Connect(ZkAddrs, time.Second)
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
	_, _= k.client.Create("/alivemeta", []byte("alivemeta"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/alivecoord", []byte("alive"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/alivefs", []byte("alive"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/data", []byte("data"), int32(0), zk.WorldACL(zk.PermAll))

	if k.coordaddr != "" && k.fsaddr != "" {
		_, err := k.client.Create("/alivecoord/"+k.coordaddr, []byte(k.coordaddr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
		_, err = k.client.Create("/alivefs/"+k.fsaddr, []byte(k.fsaddr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
    empty := map[string]string{}
    sm := ServerMeta{PrimaryFor: empty, ReplicaFor: empty}
    d, e := json.Marshal(&sm)
    if e != nil { return e }
    _, err = k.client.Create("/alivemeta/"+k.coordaddr, d, int32(0), zk.WorldACL(zk.PermAll))
    _, err = k.client.Create("/alivemeta/"+k.fsaddr, d, int32(0), zk.WorldACL(zk.PermAll))
		if err != nil {
			log.Fatalf("error creating node in zkserver")
			return err
		}
	}
	// set up serverfs, servercoords
  go k.Watch()
	return nil
}

func (k *KeeperClient) Watch() {
    for {
        backs, watch, e := k.AliveWatch()
        if e != nil { panic(e) }
        if len(backs) < len(k.serverfs) {
            k.UpdateBackends()
        } else if len(backs) > len(k.serverfs) {
            k.UpdateBackends()
        }
        <-watch
    }
}


func (k *KeeperClient) AliveWatch() ([]string, <-chan zk.Event, error) {
    backs, _, watch, e := k.client.ChildrenW("/alivecoord")
    return backs, watch, e
}

func (k *KeeperClient) GetBackendMaps() (map[string]*ClientFs, map[string]*ClientFs, error) {
	e := k.UpdateBackends()
	if e != nil { return nil, nil, e }
	coordmap := map[string]*ClientFs{}
	fsmap := map[string]*ClientFs{}
	for i, coord := range k.servercoords {
		coordmap[coord.Addr] = coord
		fsmap[k.serverfs[i].Addr] = k.serverfs[i]
	}
	return coordmap, fsmap, nil
}

func (k *KeeperClient) GetBackends() ([]*ClientFs, []*ClientFs, error) {
	fmt.Println("getting backends")
	e := k.UpdateBackends()
	if e != nil { return nil, nil, e }
	return k.servercoords, k.serverfs, nil
}

func (k *KeeperClient) UpdateBackends() error {
	coordbacks, _, _, e := k.client.ChildrenW("/alivecoord")
	if e != nil {
		log.Fatalf("error getting alive nodes", e)
		return e
	}
	fsbacks, _, _, e := k.client.ChildrenW("/alivefs")
	if e != nil {
		log.Fatalf("error getting alive nodes", e)
		return e
	}
	if len(coordbacks) != len(fsbacks) {
		return fmt.Errorf("Error with zkget")
	}

	done := make(chan bool)
	servercoords := []*ClientFs{}
	serverfs:= []*ClientFs{}

	for i, addr := range fsbacks {
		go func(a string) {
			c := NewClientFs(a)
			// if it is not this server then connect
			if a != k.fsaddr {
				e := c.Connect()
				if e != nil {
					log.Println("keeper couldnt connect to backend", a)
				}
			}
			serverfs = append(serverfs, c)
			done <- true
		}(addr)

		go func(a string) {
			c := NewClientFs(a)
			// if it is not this server then connect
			if a != k.fsaddr {
				e := c.Connect()
				if e != nil {
					log.Println("keeper couldnt connect to backend", a)
				}
			}
			servercoords = append(servercoords, c)
			done <- true
		}(coordbacks[i])
	}

	for i := 0; i < len(fsbacks) + len(coordbacks); i++ {
		<-done
	}

	if len(servercoords) != len(serverfs) {
		return fmt.Errorf("Coords didnt match ServerFS\n")
	}
	if len(servercoords) == 0 {
		return fmt.Errorf("No backends online yet\n")
	}

	k.servercoords = servercoords
	k.serverfs = serverfs
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

func (k *KeeperClient) GetWatch(watch string) (<-chan zk.Event, error) {
    _, _, w, e := k.client.ChildrenW(watch)
    fmt.Println(e)
    if e != nil { return nil, e }
    return w, e
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
        primary := ServerFileMeta{k.coordaddr, 0, 0}
        // pick a replica on the median
        replica := k.serverfs[len(k.serverfs)/2].Addr
        var kmeta KeeperMeta
        if replica != k.fsaddr {
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
        k.AddServerMeta(path, k.coordaddr, false)
        k.AddServerMeta(path, replica, true)
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
    if k.coordaddr != kmeta.Primary.Addr {
        sfm := kmeta.Replicas[k.fsaddr]
        if read {
            sfm.ReadCount += 1
        } else {
            sfm.WriteCount += 1
        }
        kmeta.Replicas[k.fsaddr] = sfm
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

// add a primary or a replica to a server's metadata
func (k *KeeperClient) AddServerMeta(path, addr string, replica bool) error {
    smeta, _, e := k.client.Get("/alivemeta/" + addr)
    if e != nil { return e}

    var sm ServerMeta
    e = json.Unmarshal(smeta, &sm)
    if e != nil { return e }
    if replica {
        if sm.ReplicaFor[path] == "" {
            sm.ReplicaFor = map[string]string{path: path}
        } else {
            sm.ReplicaFor[path] = path
        }
    } else {
        if sm.PrimaryFor[path] == "" {
            sm.PrimaryFor = map[string]string{path: path}
        } else {
            sm.PrimaryFor[path] = path
        }
    }
    smdata, e := json.Marshal(sm)
    if e != nil { fmt.Println("marshal error", e); return e }
    _, e = k.client.Set("/alivemeta/"+addr, smdata, -1)
    if e != nil { fmt.Println("set error", e) ;return e }
    return nil
}

func (k *KeeperClient) RemoveServerMeta(path, addr string, replica bool) error {
    smeta, _, e := k.client.Get("/alivemeta/" + addr)
    if e != nil { return e }

    var sm ServerMeta
    e = json.Unmarshal(smeta, &sm)
    if e != nil { return e }
    if replica {
        delete(sm.ReplicaFor, path)
    } else {
        delete(sm.PrimaryFor, path)
    }
    smdata, e := json.Marshal(&sm)
    if e != nil { return e }
    _, e = k.client.Set("/alivemeta/"+addr, smdata, -1)
    if e != nil { return e }
    return nil
}
