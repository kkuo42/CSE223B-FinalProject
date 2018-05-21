package proj

import (
    "encoding/gob"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// RPC interface input output structs
type Open_input struct {
	Name string
	Flags uint32
	Context *fuse.Context
}
type Open_output struct {
	// FuseFile nodefs.File // repaced witha FrontendFile and openFiles
	FileId int
	Status fuse.Status
}

type OpenDir_input struct {
	Name string
	Context *fuse.Context
}
type OpenDir_output struct {
	Stream []fuse.DirEntry
	Status fuse.Status
}

type GetAttr_input struct {
	Name string
	Context *fuse.Context
}
type GetAttr_output struct {
	Attr *fuse.Attr
	Status fuse.Status
}

type FileRead_input struct {
	FileId int
	Off int64
	BuffLen int
}
type FileRead_output struct {
	Dest []byte
	ReadResult fuse.ReadResult
	Status fuse.Status
}

type BackendFs interface {
	// wrappers for pathfs loopback file system calls
	Open(input *Open_input, output *Open_output) error
	OpenDir(input *OpenDir_input, output *OpenDir_output) error
	GetAttr(input *GetAttr_input, output *GetAttr_output) error

	// wrappers for file calls
	FileRead(input *FileRead_input, output *FileRead_output) error
}

type ServerFs struct {
	fs pathfs.FileSystem
	openFiles []nodefs.File
} 
func NewServerFs(directory string) ServerFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
    fs := NewCustomLoopbackFileSystem(directory)
	return ServerFs{fs: fs}
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

func (self *ServerFs) FileRead(input *FileRead_input, output *FileRead_output) error {
	output.Dest = make([]byte, input.BuffLen) // recreates the buffer on server for client/server or replaces orignal for local
	output.ReadResult, output.Status = self.openFiles[input.FileId].Read(output.Dest, input.Off)
	return nil
}


// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)