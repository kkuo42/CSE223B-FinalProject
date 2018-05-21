package proj

import (
	"github.com/hanwen/go-fuse/fuse"
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


