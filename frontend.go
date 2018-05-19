package proj

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)
/*
the frontend that the client uses with FUSE

it is initialized with pathfs.NewDefaultFileSystem() in fs-client so that it 
implements a FileSystem that returns ENOSYS for every operation that is not 
overrided
*/
type Frontend struct {
	pathfs.FileSystem
	backendFs BackendFs
}
func NewFrontendRemotelyBacked(addr string) Frontend {
 	fs := pathfs.NewDefaultFileSystem()
    clientFs := ClientFs{addr: addr}
    clientFs.Connect()
    return Frontend{FileSystem: fs, backendFs: &clientFs}
}
func NewFrontendLocalyBacked(directory string) Frontend {
 	fs := pathfs.NewDefaultFileSystem()
    serverFs := NewServerFs(directory)
    return Frontend{FileSystem: fs, backendFs: &serverFs}
}
func (self *Frontend) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	input := &Open_input{name: name, flags: flags, context: context}
	output := &Open_output{}
	
	e := self.backendFs.Open(input, output)
	
	if e != nil {
		return nil, fuse.ENOSYS // probably shoud have different error handling for rpc fail
	}
	
	return output.fuseFile, output.status
}
func (self *Frontend) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	input := &OpenDir_input{name: name, context: context}
	output := &OpenDir_output{}
	
	e := self.backendFs.OpenDir(input, output)
	
	if e != nil {
		return nil, fuse.ENOSYS // probably shoud have different error handling for rpc fail
	}
	
	return output.stream, output.status
}
