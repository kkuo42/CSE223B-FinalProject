package proj

import (
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

// TODO: (maybe?) implement a custom pathfs.FileSystem, fs is currently initialized with a loopback in fs-server/main.go
// check https://github.com/hanwen/go-fuse/blob/master/fuse/pathfs/api.go for interface to implement
type ServerFs struct {
	fs pathfs.FileSystem
	openFiles []nodefs.File
} 
func NewServerFs(directory string) ServerFs {
	// register used datatype
    fs := pathfs.NewLoopbackFileSystem(directory)
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
	output.ReadResult, output.Status = self.openFiles[input.FileId].Read(output.Dest, input.Off)
	return nil
}


// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)



type CustomLoopbackFileSystem struct {
	pathfs.FileSystem
}
func NewCustomLoopbackFileSystem(directory string) CustomLoopbackFileSystem {
	return CustomLoopbackFileSystem{FileSystem: pathfs.NewLoopbackFileSystem(directory)}
}
// overide read so that it does NOT use ReadResultFd but uses ReadResultData
// may need to rewrite the entire thing because do not have access to the loop backs lock
func (f *CustomLoopbackFileSystem) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	// f.lock.Lock()

	sz := len(buf)
	if len(buf) < sz {
		sz = len(buf)
	}

	n, err := syscall.Pread(int(r.Fd), buf[:sz], off)
	if err == io.EOF {
		err = nil
	}

	if n < 0 {
		n = 0
	}

	r := ReadResultData(buf[:n])

	// f.lock.Unlock()
	return r, ToStatus(err)
}