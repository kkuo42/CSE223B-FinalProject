package proj

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// RPC interface
type Open_input struct {
	name string
	flags uint32
	context *fuse.Context
}
type Open_output struct {
	fuseFile nodefs.File
	status fuse.Status
}

type OpenDir_input struct {
	name string
	context *fuse.Context
}
type OpenDir_output struct {
	stream []fuse.DirEntry
	status fuse.Status
}

type BackendFs interface {
	Open(input *Open_input, output *Open_output) error
	OpenDir(input *OpenDir_input, output *OpenDir_output) error
}

// TODO: implement a custom pathfs.FileSystem, fs is currently initialized with a loopback in fs-server/main.go
// check https://github.com/hanwen/go-fuse/blob/master/fuse/pathfs/api.go for interface to implement
type ServerFs struct {
	fs pathfs.FileSystem
} 
func NewServerFs(directory string) ServerFs {
    fs := pathfs.NewLoopbackFileSystem(directory)
	return ServerFs{fs: fs}
}
func (self *ServerFs) Open(input *Open_input, output *Open_output) error { 
	output.fuseFile, output.status = self.fs.Open(input.name, input.flags, input.context)
	return nil
}
func (self *ServerFs) OpenDir(input *OpenDir_input, output *OpenDir_output) error { 
	output.stream, output.status = self.fs.OpenDir(input.name, input.context)
	return nil
}

// assert that ServerFs implements BackendFs
var _ BackendFs = new(ServerFs)
