package proj

import (
	"fmt"
	"encoding/gob"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type ServerFs struct {
	addr string
	fs pathfs.FileSystem
	kc *KeeperClient
	openFiles []nodefs.File
}

func NewServerFs(directory, addr, zkaddr string) ServerFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	fs := NewCustomLoopbackFileSystem(directory)
	kc := NewKeeperClient(zkaddr, addr)
	e := kc.Init()
	if e != nil {
		panic(e)
	}

	return ServerFs{addr: addr, fs: fs, kc: kc}
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
	entries, e := self.kc.GetChildrenAttributes(input.Name)
	if e != nil {
		return e
	}

	output.Stream = entries
	return nil
}

func (self *ServerFs) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	// fetch the attr from zk
	kmeta, e := self.kc.Get(input.Name)
	if e != nil {
		// do nothing
	}
	output.Attr = &kmeta.Attr
	if output.Attr.Ino == 0 {
		output.Attr, output.Status = self.fs.GetAttr(input.Name, input.Context)
	}

	return nil
}

func (self *ServerFs) Rename(input *Rename_input, output *Rename_output) error {
	output.Status = self.fs.Rename(input.Old, input.New, input.Context)
	fmt.Println(output.Status)
	a, _ := self.fs.GetAttr(input.New, input.Context)

	e := self.kc.Create(input.New, *a)
	if e != nil {
		// do nothing for now
		fmt.Println("mv error", e)
		return e
	}

	e = self.kc.Remove(input.Old)
	if e != nil {
		// do nothing for now
		fmt.Println("mv error", e)
		return e
	}
	return nil
}

func (self *ServerFs) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	output.Status = self.fs.Mkdir(input.Name, input.Mode, input.Context)
	// get attributes after make 
	a, _ := self.fs.GetAttr(input.Name, input.Context)

	e := self.kc.Create(input.Name, *a)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerFs) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	output.Status = self.fs.Rmdir(input.Name, input.Context)
	e := self.kc.RemoveDir(input.Name)

	if e != nil {
		return e
	}

	return nil
}

func (self *ServerFs) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	output.Status = self.fs.Unlink(input.Name, input.Context)
	e := self.kc.Remove(input.Name)

	if e != nil {
		return e
	}

	return nil
}

func (self *ServerFs) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)
	loopbackFile, status := self.fs.Create(input.Path, input.Flags, input.Mode, input.Context)
	a, _ := self.fs.GetAttr(input.Path, input.Context)

	// TODO after create check if success or fail
	e := self.kc.Create(input.Path, *a)
	if e != nil {
		return e
	}
	self.openFiles = append(self.openFiles, loopbackFile)
	output.FileId = len(self.openFiles)-1
	output.Status = status
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
	fmt.Println("fs write")

	// after we have written the file we will go an update the node that we created/modified
	a, _ := self.fs.GetAttr(input.Path, input.Context)
	kmeta, e := self.kc.Get(input.Path)
	if e != nil {
		return e
	}
	fmt.Println("write get")

	kmeta.Attr = *a
	e = self.kc.Set(input.Path, kmeta)
	if e != nil {
		return e
	}
	fmt.Println("end of write")
	return nil
}

// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)
