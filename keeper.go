package proj

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/samuel/go-zookeeper/zk"
	"encoding/json"
	"time"
	"log"
	"fmt"
	"math/rand"
	"errors"
	"strings"
)

type ServerMeta struct {
	PrimaryFor map[string]string
	ReplicaFor map[string]string
}

type ServerFileMeta struct {
	CoordAddr string
	SFSAddr string
	WriteCount int
	ReadCount int
}

// this struct is to maintain all information relative to the keeper
type KeeperMeta struct {
	Attr fuse.Attr
	Deleted bool
	Primary ServerFileMeta
	Replicas map[string]ServerFileMeta
}

type KeeperClient struct {
	coordaddr string
	fsaddr string
	client *zk.Conn
	serverfs []*ClientFs
	servercoords []*ClientFs
	latencyMap map[string]int
}

func NewKeeperClient(coordaddr, fsaddr string) *KeeperClient {
	return &KeeperClient{coordaddr: coordaddr, fsaddr: fsaddr, latencyMap: make(map[string]int)}
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
		_, err := k.client.Create("/alivecoord/"+k.coordaddr+"_", []byte(k.coordaddr), SequentialEphemeral, zk.WorldACL(zk.PermAll))
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
        _, watch, e := k.AliveWatch()
        if e != nil { panic(e) }
	// always keep backs up to date in keeper
	k.UpdateBackends()
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

type connectRes struct {
	back *ClientFs
	latency int
}

// Used to get backends for frontend to connect to with artificial random latency added
func (k *KeeperClient) GetBackendForFrontend(pref string) (*ClientFs, int, error) {
	fmt.Println("updating backs")
	coordbacks, _, _, e := k.client.ChildrenW("/alivecoord")
	if e != nil {
		log.Fatalf("error getting alive nodes")
		return nil, 0, e
	}

	// Designate random delay for coordinators
	for _, addr := range coordbacks {
		_, ok := k.latencyMap[addr]
		if !ok {
			k.latencyMap[addr] = rand.Intn(Latency)
		}
	}

	// if the coord has a preference assign it that server if possible
	if pref != "" {
		fmt.Println("SERVER HAS A PREF")
		for _, addr := range coordbacks {
			saddr := strings.Split(addr, "_")[0]
			if pref == saddr {
				c := NewClientFs(addr)
				e := c.Connect()
				if e != nil {
					fmt.Println("kc couldnt connect to pref back", addr)
				} else {
					return c, k.latencyMap[addr], nil
				}
			}
		}
	}

	done := make(chan connectRes)
	for _, addr := range coordbacks {

		go func(coordaddr string) {
			c := NewClientFs(coordaddr)
			e := c.Connect()
			if e != nil {
				fmt.Println("kc couldnt connect to coordinator", c.Addr)
				return
			}
			time.Sleep(time.Millisecond * time.Duration(k.latencyMap[coordaddr]))
			fmt.Println("kc connected after fake latency of", k.latencyMap[coordaddr],"to coordinator", coordaddr)
			done <- connectRes{c, k.latencyMap[coordaddr]}
		}(addr)
	}

	result := <-done

	return result.back, result.latency, nil
}

func (k *KeeperClient) UpdateBackends() error {
	fmt.Println("updating backs")
	coordbacks, _, _, e := k.client.ChildrenW("/alivecoord")
	if e != nil {
		log.Fatalf("error getting alive nodes")
		return e
	}

	fsbacks, _, _, e := k.client.ChildrenW("/alivefs")
	if e != nil {
		log.Fatalf("error getting alive nodes")
		return e
	}
	if len(coordbacks) != len(fsbacks) {
		return fmt.Errorf("Error with zkget")
	}

	done := make(chan bool)
	servercoords := []*ClientFs{}
	serverfs:= []*ClientFs{}

	for i, addr := range fsbacks {
		go func(fsaddr string) {
			c := NewClientFs(fsaddr)
			e := c.Connect()
			if e != nil {
				log.Println("kc couldnt connect to serverfs", fsaddr)
			}
			serverfs = append(serverfs, c)
			done <- true
		}(addr)
		go func(coordaddr string) {
			c := NewClientFs(coordaddr)
			e := c.Connect()
			if e != nil {
				log.Println("kc couldnt connect to coordinator", coordaddr)
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

func (k *KeeperClient) Get(path string) ([]byte, error) {
	data, _, e := k.client.Get(path)
	return data, e
}

func (k *KeeperClient) GetData(path string) (KeeperMeta, error) {
	data, e := k.Get("/data/" + path)
	if e != nil {
		// do nothing for now, should crash
		return KeeperMeta{}, e
	}
	var kmeta KeeperMeta
	e = json.Unmarshal(data, &kmeta)
	if kmeta.Deleted {
		return kmeta, errors.New("Deleted boolean")
	}
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

func (k *KeeperClient) Children(path string) ([]string, error) {
	children, _, e := k.client.Children(path)
	if e != nil { return nil, e }
	return children, e
}

func (k *KeeperClient) GetChildren(path string) ([]string, error) {
	// if in the root directory don't add /
	inputstr := "/data"
	if path != "" {
		inputstr += "/"+path
	}

	return k.Children(inputstr)
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
		fm, e := k.GetData(p)
		if e != nil && e.Error() != "Deleted boolean"{
			log.Println("error here:", e)
			return nil, e
		}
		if !fm.Deleted{
			fileEntries = append(fileEntries, fuse.DirEntry{Name: f, Mode: fm.Attr.Mode})
		}
	}
	return fileEntries, nil
}

func (k *KeeperClient) Create(path string, attr fuse.Attr, deleted bool) (string, error) {
	data, e := k.Get("/data/"+path)
	var kmeta KeeperMeta
	if e == nil {
		e = json.Unmarshal(data, &kmeta)
		if kmeta.Deleted {
			kmeta.Attr = attr
			kmeta.Deleted = false
			d, _ := json.Marshal(&kmeta)
			_, e = k.client.Set("/data/"+path, []byte(d), -1)
		} else {
			return "", errors.New("value already exists in keeper, but isn't deleted")
		}
	} else {

	    // pick a replica on the median
	    replicaAddr := k.serverfs[len(k.serverfs)/2].Addr
	    if replicaAddr == k.coordaddr && len(k.serverfs) > 1 {
		    // you picked yourself
		    replicaAddr = k.serverfs[len(k.serverfs)/2-1].Addr
	    }

		if Debug {
			for _, replica := range k.serverfs {
				if ReplicaAddrs[k.fsaddr] == replica.Addr {
					// safe to assign
					replicaAddr = ReplicaAddrs[k.fsaddr]
				}
			}
		}

	    // brand new file, initialize new file metadata
		serverFileMeta := ServerFileMeta{k.coordaddr, k.fsaddr, 0, 1}

		fmt.Printf("assigning server %v as replica\n", replicaAddr)
		replicaAddrs := map[string]ServerFileMeta{}
		var kmeta KeeperMeta
		if replicaAddr != k.fsaddr { //THIS SHOULD ALWAYS BE TRUE?
			var replicaCoordAddr string
			addr_type := "coord"
			// find the serverfs with the right addr
			index := -1
			for i, temp := range k.serverfs {
				if temp.Addr == replicaAddr {
					index = i
					break
				}
			}
			if index != -1 {
				e := k.serverfs[index].GetAddress(&addr_type, &replicaCoordAddr)
				if e != nil {
					panic(e)
				}
			}
			replicaAddrs[replicaAddr] = ServerFileMeta{replicaCoordAddr, replicaAddr, 0, 0}
		}
		kmeta = KeeperMeta{Primary: serverFileMeta, Replicas: replicaAddrs, Attr: attr}
		d, e := json.Marshal(&kmeta)
		if e != nil {
			return "", e
		}

		_, e = k.client.Create("/data/"+path, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
		if e != nil {
			return "", e
		}
	    k.AddServerMeta(path, k.coordaddr, false)

		if replicaAddr != k.fsaddr {
			k.AddServerMeta(path, replicaAddr, true)
			return replicaAddr, nil
		}
	}
	return "", nil
}

func (k *KeeperClient) Delete(path string) error {
	e := k.client.Delete(path, -1)
	if e != nil { return e }
	return nil
}

func (k *KeeperClient) Remove(path string, data KeeperMeta) error {
	data.Deleted = true
	// remove all of the servers that had this path
	k.RemoveServerMeta(path, data.Primary.CoordAddr, false)
	for addr, _ := range data.Replicas {
		k.RemoveServerMeta(path, addr, true)
	}
	m, e := json.Marshal(&data)
	_, e = k.client.Set("/data/"+path, m, -1)
	if e != nil {
		return e
	}

	return nil
}

func (k *KeeperClient) RemoveDir(path string) error {
	// recursively delete all children
	kmeta, e := k.GetData(path)
	children, e := k.GetChildren(path)
	if len(children) == 0 || e != nil {
		// it is safe to delete this node
		k.Remove(path, kmeta)
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

func (k *KeeperClient) IncAddr(path, addr string, kmeta KeeperMeta, read bool) error {
	if addr != kmeta.Primary.CoordAddr {
		if sfm, ok := kmeta.Replicas[addr]; ok {
			if read {
			    sfm.ReadCount += 1
			} else {
			    sfm.WriteCount += 1
			}
			kmeta.Replicas[addr] = sfm
		}
	} else {
		if read {
		    kmeta.Primary.ReadCount += 1
		} else {
		    kmeta.Primary.WriteCount += 1
		}
	}
	e := k.Set(path, kmeta)
	if e != nil { return e }
	return nil
}

func (k *KeeperClient) Inc(path string, read bool) error {
	// if this server is not the primary go increment in replica
	kmeta, e := k.GetData(path)
	if e != nil { return e }
	if k.coordaddr != kmeta.Primary.CoordAddr {
		return k.IncAddr(path, k.fsaddr, kmeta, read)
	} else {
		return k.IncAddr(path, k.coordaddr, kmeta, read)
	}
}

// add a primary or a replica to a server's metadata
func (k *KeeperClient) AddServerMeta(path, addr string, replica bool) error {
    smeta, e := k.Get("/alivemeta/" + addr)
    if e != nil { return e}

    var sm ServerMeta
    e = json.Unmarshal(smeta, &sm)
    if e != nil { return e }
    if replica {
            sm.ReplicaFor[path] = path
    } else {
            sm.PrimaryFor[path] = path
    }
    smdata, e := json.Marshal(sm)
    if e != nil { fmt.Println("marshal error", e); return e }
    _, e = k.client.Set("/alivemeta/"+addr, smdata, -1)
    if e != nil { fmt.Println("set error", e) ;return e }
    return nil
}

func (k *KeeperClient) RemoveServerMeta(path, addr string, replica bool) error {
    smeta, e := k.Get("/alivemeta/" + addr)
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
