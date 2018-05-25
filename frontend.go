package proj

import (
    "log"
	"time"
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)
/*
the frontend that the client uses with go-fuse

it is initialized with pathfs.NewDefaultFileSystem() in fs-client so that it 
implements a FileSystem that returns ENOSYS for every operation that is not 
overrided

check https://github.com/hanwen/go-fuse/blob/master/fuse/pathfs/api.go for 
interface to implement
*/

type Frontend struct {
	pathfs.FileSystem
	backendFs BackendFs
        kc *KeeperClient
}

func NewFrontendRemotelyBacked(backaddr string) Frontend {
    fs := pathfs.NewDefaultFileSystem()

    // "" indicates this is a frontend and to not add to /alive
    kc := NewKeeperClient("")
    e := kc.Init()
    if e != nil { panic(e) }

    backends, e := kc.GetBackends()
    // TODO watch alive for changes in state and update list of clients
    if e != nil {
	    log.Fatalf("error getting alive nodes: %v\n", e)
    }
    if len(backends) == 0 {
	    log.Fatalf("ERROR: no backends")
    }

    // get the closest backend
    clientFs := backends[0]
    if backaddr != "" {
	    clientFs = NewClientFs(backaddr)
    }
    e = clientFs.Connect()
    if e != nil {
		log.Fatalf("error connecting to backend. try another one?, error: %v",e)
    }
    log.Println("Connected to backend:", clientFs.Addr)
    return Frontend{FileSystem: fs, backendFs: clientFs}
}

func (self *Frontend) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	fmt.Println("Open:", name)
	input := &Open_input{Name: name, Flags: flags, Context: context}
	output := &Open_output{}

	e := self.backendFs.Open(input, output)

	if e != nil {
		log.Fatalf("Fuse call to backendFs.Open failed: %v\n%v, %v\n", e, output.FileId, output.Status)
		// return nil, fuse.ENOSYS // probably shoud have different error handling for rpc fail
	}

	fuseFile = &FrontendFile{FileId: output.FileId, Backend: self.backendFs, Name: name, Context: context}

	return fuseFile, output.Status
}
func (self *Frontend) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	input := &OpenDir_input{Name: name, Context: context}
	output := &OpenDir_output{}

	e := self.backendFs.OpenDir(input, output)

	if e != nil {
	    log.Fatalf("Fuse call to backendFs.OpenDir failed: %v\n", e)
		// return nil, fuse.ENOSYS // probably shoud have different error handling for rpc fail
	}

	return output.Stream, output.Status
}
func (self *Frontend) GetAttr(name string, context *fuse.Context) (attr *fuse.Attr, status fuse.Status) {
	input := &GetAttr_input{Name: name, Context: context}
	output := &GetAttr_output{}

	e := self.backendFs.GetAttr(input, output)

	if e != nil {
	    log.Fatalf("Fuse call to backendFs.GetAttr failed: %v\n", e)
		// return nil, fuse.ENOSYS // probably shoud have different error handling for rpc fail
	}

	return output.Attr, output.Status
}

func (self *Frontend) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	fmt.Println("Unlink:",name)
	input := &Unlink_input{name, context}
	output := &Unlink_output{}

	e := self.backendFs.Unlink(input, output)

	if e != nil {
		return fuse.ENOSYS
	}

	return output.Status
}

func (self *Frontend) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	input := &Rename_input{Old: oldName, New: newName, Context: context}
	output := &Rename_output{}

	e := self.backendFs.Rename(input, output)

	if e != nil {
		return fuse.ENOSYS
	}
	return output.Status
}

func (self *Frontend) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	input := &Mkdir_input{Name: name, Mode: mode, Context: context}
	output := &Mkdir_output{}

	e := self.backendFs.Mkdir(input, output)

	if e != nil {
		return fuse.ENOSYS
	}
	return output.Status
}

func (self *Frontend) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	input := &Rmdir_input{Name: name, Context: context}
	output := &Rmdir_output{}

	e := self.backendFs.Rmdir(input, output)

	if e != nil {
		return fuse.ENOSYS
	}
	return output.Status
}

func (self *Frontend) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	input := &Create_input{path, flags, mode, context}
	output := &Create_output{}

	e := self.backendFs.Create(input, output)

	if e != nil {
		return nil, fuse.ENOSYS
	}
	fuseFile = &FrontendFile{FileId: output.FileId, Backend: self.backendFs, Name: path, Context: context}

	return fuseFile, output.Status
}



// A frontend file is passed to the fuse front end, it has the means to forward the operations to a file on the backed server
type FrontendFile struct {
	Name string
	FileId int
	Backend BackendFs
	Context *fuse.Context
}
func (self *FrontendFile) SetInode(*nodefs.Inode) {} //ok
func (self *FrontendFile) String() string {return fmt.Sprintf("FrontendFile(%v:%v)", self.Backend, self.FileId)}
func (self *FrontendFile) InnerFile() nodefs.File {return nil} //ok

func (self *FrontendFile) Read(dest []byte, off int64) (readResult fuse.ReadResult, status fuse.Status) {
	fmt.Println("Read:", self.Name)
	input := &FileRead_input{Path: self.Name, FileId: self.FileId, Off: off, BuffLen: len(dest)}
	output := &FileRead_output{Dest: dest, ReadResult: readResult, Status: status}
	e := self.Backend.FileRead(input, output)
	if e != nil {
	    // log.Fatalf("backend faild to read file: %v\n", e)
		return nil, fuse.EIO
	}
	return output.ReadResult, output.Status
}

func (self *FrontendFile) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	input := &FileWrite_input{self.Name, self.FileId, data, off, self.Context}
	output := &FileWrite_output{}
	e := self.Backend.FileWrite(input, output)

	if e != nil {
		return 0, fuse.ENOSYS
	}

	return output.Written, output.Status
}

func (self *FrontendFile) Flock(flags int) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Flush() fuse.Status {return fuse.ENOSYS} //TODO?

func (self *FrontendFile) Release() {
	input := &FileRelease_input{self.Name, self.FileId}
	output := &FileRelease_output{}
	self.Backend.FileRelease(input, output)
}

func (self *FrontendFile) Fsync(flags int) (code fuse.Status) {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Truncate(size uint64) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) GetAttr(out *fuse.Attr) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Chown(uid uint32, gid uint32) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Chmod(perms uint32) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Allocate(off uint64, size uint64, mode uint32) (code fuse.Status) {return fuse.ENOSYS} //TODO?

var _ nodefs.File = new(FrontendFile)
