package proj

import (
	"fmt"
	"time"
	"strings"
    "encoding/gob"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/samuel/go-zookeeper/zk"
)

type ServerFs struct {
	addr string
	fs pathfs.FileSystem
	zkClient *zk.Conn
	openFiles []nodefs.File
}

func NewServerFs(directory string, addr string, zkaddr string) ServerFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
    fs := NewCustomLoopbackFileSystem(directory)

	zkClient, _, err := zk.Connect(strings.Split(zkaddr, ","), time.Second)
	// Just panic for now, should fix later
	if err != nil {
		panic(err)
	}

	return ServerFs{addr: addr, fs: fs, zkClient: zkClient}
}

func (self *ServerFs) Open(input *Open_input, output *Open_output) error {

	loopbackFile, status := self.fs.Open(input.Name, input.Flags, input.Context)
	self.openFiles = append(self.openFiles, loopbackFile)
	output.FileId = len(self.openFiles)-1
	output.Status = status
	return nil
}

func (self *ServerFs) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	output.Stream, output.Status = self.fs.OpenDir(input.Name, input.Context)
	return nil
}

func (self *ServerFs) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	output.Attr, output.Status = self.fs.GetAttr(input.Name, input.Context)
	return nil
}

func (self *ServerFs) Rename(input *Rename_input, output *Rename_output) error {
	output.Status = self.fs.Rename(input.Old, input.New, input.Context)
	return nil
}

func (self *ServerFs) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	output.Status = self.fs.Mkdir(input.Name, input.Mode, input.Context)
	return nil
}

func (self *ServerFs) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	output.Status = self.fs.Rmdir(input.Name, input.Context)
	return nil
}

func (self *ServerFs) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	output.Status = self.fs.Unlink(input.Name, input.Context)

	err := self.zkClient.Delete("/"+input.Name, -1)

	if err != nil {
		panic(err)
	}
	return nil
}

func (self *ServerFs) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create: "+input.Path)
	loopbackFile, status := self.fs.Create(input.Path, input.Flags, input.Mode, input.Context)
	_, err := self.zkClient.Create("/"+input.Path, []byte(self.addr), int32(0), zk.WorldACL(zk.PermAll))

	if err != nil {
		panic(err)
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
	output.Written, output.Status = self.openFiles[input.FileId].Write(input.Data, input.Off)
	return nil
}

// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)
