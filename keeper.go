package proj

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/samuel/go-zookeeper/zk"
	"encoding/json"
	"strings"
	"time"
	"log"
)

// this struct is to maintain all information relative to the keeper
type KeeperMeta struct {
	Attr fuse.Attr
	Primary string
	Replicas []string
}

type KeeperClient struct {
	zkaddr string
	pubaddr string
	client *zk.Conn
	backends []string
}

type KeeperHandler struct {
	// TODO, when servers go down/up/move 
}

func NewKeeperClient(addr, pubaddr string) *KeeperClient {
	return &KeeperClient{zkaddr: addr, pubaddr: pubaddr}
}

func (k *KeeperClient) Init() error {
	client, _, err := zk.Connect(strings.Split(k.zkaddr, ","), time.Second)
	k.client = client

	if err != nil {
		log.Fatalf("error connecting to zkserver\n")
		return err
	}

	// attempt to create alive and data dirs, if it fails itll be caught below
	_, _= k.client.Create("/alive", []byte("alive"), int32(0), zk.WorldACL(zk.PermAll))
	_, _= k.client.Create("/data", []byte("data"), int32(0), zk.WorldACL(zk.PermAll))

	// if there is no error then we want to register that this server is alive
	_, err = k.client.Create("/alive/"+k.pubaddr, []byte(k.pubaddr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Fatalf("error creating node in zkserver")
		return err
	}
	return nil
}

func (k *KeeperClient) Get(path string) (KeeperMeta, error) {
	data, _, e := k.client.Get("/data/" + path)
	if e != nil {
		// do nothing for now, should crash
		return KeeperMeta{}, e
	}
	var kmeta KeeperMeta
	e = json.Unmarshal(data, kmeta)
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
	log.Println("not get children")
	if e != nil {
		return nil, e
	}
	fileEntries := []fuse.DirEntry{}

	// for each of the files fetched get their metadata (may be unecessary)
	for _, f := range files {
		log.Println(path, f)
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
	kmeta := KeeperMeta{Primary: k.pubaddr, Attr: attr}
	d, e := json.Marshal(&kmeta)
	if e != nil {
		return e
	}

	_, e = k.client.Create("/data/"+path, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
	if e != nil {
		return e
	}
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
