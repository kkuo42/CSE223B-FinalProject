package proj

import (
	"fmt"
	"log"
	"time"
	"strings"
	"encoding/gob"
	"encoding/json"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/samuel/go-zookeeper/zk"
)

type ServerFs struct {
	addr string
	pubaddr string
	fs pathfs.FileSystem
	zkClient *zk.Conn
	openFiles []nodefs.File
}

func NewServerFs(directory, addr, pubaddr, zkaddr string) ServerFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	fs := NewCustomLoopbackFileSystem(directory)

	zkClient, _, err := zk.Connect(strings.Split(zkaddr, ","), time.Second)
	// Just panic for now, should fix later
	if err != nil {
		log.Fatalf("error connecting to zkserver\n")
		panic(err)
	}

	// if there is no error then we want to register that this server is alive
	_, e := zkClient.Create("/alive/"+pubaddr, []byte(pubaddr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if e != nil {
		log.Fatalf("error creating node in zkserver")
		panic(e)
	}

	return ServerFs{addr: addr, pubaddr: pubaddr, fs: fs, zkClient: zkClient}
}

func (self *ServerFs) Open(input *Open_input, output *Open_output) error {
	fmt.Println("opening:", input.Name)
	loopbackFile, status := self.fs.Open(input.Name, input.Flags, input.Context)
	self.openFiles = append(self.openFiles, loopbackFile)
	output.FileId = len(self.openFiles)-1
	output.Status = status
	return nil
}

func (self *ServerFs) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	// use the keeper to list all the files in the directory
	fmt.Println("opening dir:", input.Name)
	output.Stream, output.Status = self.fs.OpenDir(input.Name, input.Context)
	files, _, e := self.zkClient.Children("/data")

	fmt.Println(files)
	if e != nil {
		panic(e)
	}

	// add the keeper files to the output stream
	fileEntries := []fuse.DirEntry{}
	for _, f := range files {
		fileEntries = append(fileEntries, fuse.DirEntry{Name: f})
	}

	output.Stream = fileEntries

	return nil
}

func (self *ServerFs) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	output.Attr, output.Status = self.fs.GetAttr(input.Name, input.Context)
	if output.Attr == nil {
		// fetch the attr from zk
		data, _, e := self.zkClient.Get("/data/" + input.Name)
		if e != nil {
			// do nothing for now, should crash
		}
		var keeperdata Keeper
		e = json.Unmarshal(data, &keeperdata)
		output.Attr = &keeperdata.Attr
	}
	fmt.Println(input.Name, output.Attr)
	return nil
}

func (self *ServerFs) Rename(input *Rename_input, output *Rename_output) error {
	output.Status = self.fs.Rename(input.Old, input.New, input.Context)
	fmt.Println(output.Status)
	a, _ := self.fs.GetAttr(input.New, input.Context)

	keeperdata := Keeper{Attr: *a, Primary: self.pubaddr}
	d, _ := json.Marshal(&keeperdata)

	_, e := self.zkClient.Create("/data/"+input.New, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
	if e != nil {
		// do nothing for now
		fmt.Println("mv error", e)
		return nil
	}

	e = self.zkClient.Delete("/data/"+input.Old, -1)
	return nil
}

func (self *ServerFs) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	output.Status = self.fs.Mkdir(input.Name, input.Mode, input.Context)
	// get attributes after make 
	a, _ := self.fs.GetAttr(input.Name, input.Context)

	keeperdata := Keeper{Attr: *a, Primary: self.pubaddr}
	d, _ := json.Marshal(&keeperdata)

	_, e := self.zkClient.Create("/data/"+input.Name, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerFs) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	output.Status = self.fs.Rmdir(input.Name, input.Context)

	e := self.zkClient.Delete("/data/"+input.Name, -1)

	if e != nil {
		panic(e)
	}
	return nil
}

func (self *ServerFs) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	output.Status = self.fs.Unlink(input.Name, input.Context)

	err := self.zkClient.Delete("/data/"+input.Name, -1)

	if err != nil {
		panic(err)
	}
	return nil
}

func (self *ServerFs) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)
	// before creation check if the file already exists, if it does
	// then we should move that file to this server and TODO add as secondary
	data, _, e := self.zkClient.Get("/data/"+input.Path)
	// if the keeper returns an error the node doesnt exist
	if e != nil {
		keeperfile := Keeper{Primary: self.pubaddr}
		d, e := json.Marshal(&keeperfile)
		s, e := self.zkClient.Create("/data/"+input.Path, []byte(d), int32(0), zk.WorldACL(zk.PermAll))
		loopbackFile, status := self.fs.Create(input.Path, input.Flags, input.Mode, input.Context)
		fmt.Println("create res:",s)

		if e != nil {
			panic(e)
		}

		self.openFiles = append(self.openFiles, loopbackFile)
		output.FileId = len(self.openFiles)-1
		output.Status = status
	} else {
		// the node exists so we will go contact that server
		fmt.Println(data)
		return fmt.Errorf("cannot create a file that doesnt exist, try to read it now\n")
	}
	return nil
}

func (self *ServerFs) FileRead(input *FileRead_input, output *FileRead_output) error {
	output.Dest = make([]byte, input.BuffLen) // recreates the buffer on server for client/server or replaces orignal for local
	output.ReadResult, output.Status = self.openFiles[input.FileId].Read(output.Dest, input.Off)
	return nil
}

func (self *ServerFs) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	fmt.Println("Write:", input.FileId)
	output.Written, output.Status = self.openFiles[input.FileId].Write(input.Data, input.Off)

	a, _ := self.fs.GetAttr(input.Path, input.Context)

	// after we have written the file we will go an update the node that we created/modified
	data, _, _ := self.zkClient.Get("/data/"+input.Path)
	var keeperfile Keeper
	_ = json.Unmarshal(data, &keeperfile)
	keeperfile.Attr = *a
	m, _ := json.Marshal(&keeperfile)
	_, _ = self.zkClient.Set("/data/"+input.Path, m, -1)
	return nil
}

// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)
