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
  backpref string
}

func NewFrontendRemotelyBacked(backaddr string) Frontend {
    fs := pathfs.NewDefaultFileSystem()
    return Frontend{FileSystem: fs, backpref: backaddr}
}

func (self *Frontend) Init() {
    // "" indicates this is a frontend and to not add to /alive
    self.kc = NewKeeperClient("", "")
    e := self.kc.Init()
    if e != nil { panic(e) }

	// force initial refresh of server connections
	self.RefreshClient()
    go self.WatchBacks()
}

func (self *Frontend) WatchBacks() {
    for {
        watch, e := self.kc.GetWatch("/alivecoord")
        fmt.Println("change in servers happened... picking a new one")
        if e != nil { panic(e) }
        self.RefreshClient()
        <-watch
    }
}

func (self *Frontend) RefreshClient() error {
    backends, _, e := self.kc.GetBackends()
    if e != nil { return e }
    if self.backpref != "" {
        // attempt to connect to preferred back
        back := NewClientFs(self.backpref)
        e = back.Connect()
        if e == nil {
            self.backendFs = back
            fmt.Println("pref connected to", back.Addr)
            return nil
        }
        fmt.Println("couldnt connect to prefered back")
    }
    for _, back := range backends {
        e = back.Connect()
        if e == nil {
            self.backendFs = back
            fmt.Println("connected to", back.Addr)
            return nil
        }
    }
	fmt.Println("Didn't connect to any backend")
    return e
}

func (self *Frontend) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	fmt.Println("Open:", name)
	input := &Open_input{Name: name, Flags: flags, Context: context}
	output := &Open_output{}

	e := self.backendFs.Open(input, output)

	if e != nil {
		log.Fatalf("Fuse call to backendFs.Open failed: %v\n, %v\n, we will find a new server", e, output.Status)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Open(name, flags, context)
	}
	if output.Status == fuse.OK {
		fuseFile = &FrontendFile{Name: name, Backend: self.backendFs, Context: context}
	}

	return fuseFile, output.Status
}

func (self *Frontend) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	fmt.Println("opening dir:", name)
	input := &OpenDir_input{Name: name, Context: context}
	output := &OpenDir_output{}

	e := self.backendFs.OpenDir(input, output)

	if e != nil {
    log.Fatalf("Fuse call to backendFs.OpenDir failed: %v\n, find new server", e)
    e = self.RefreshClient()
    if e != nil { panic(e) }
    return self.OpenDir(name, context)
	}

	return output.Stream, output.Status
}

func (self *Frontend) GetAttr(name string, context *fuse.Context) (attr *fuse.Attr, status fuse.Status) {
	fmt.Println("get attr:", name)

	input := &GetAttr_input{Name: name, Context: context}
	output := &GetAttr_output{}

	if self.backendFs == nil {
		fmt.Println("backendFs is nil")
	}
	e := self.backendFs.GetAttr(input, output)

	if e != nil {
    log.Fatalf("Fuse call to backendFs.GetAttr failed: %v\n", e)
    e = self.RefreshClient()
    if e != nil { panic(e) }
    return self.GetAttr(name, context)
	}

	return output.Attr, output.Status
}

func (self *Frontend) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	fmt.Println("Unlink:",name)
	input := &Unlink_input{name, context}
	output := &Unlink_output{}

	e := self.backendFs.Unlink(input, output)

	if e != nil {
                log.Fatalf("Fuse call to backendFs.Unlink failed: %v\n", e)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Unlink(name, context)
	}

	return output.Status
}

func (self *Frontend) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	input := &Rename_input{Old: oldName, New: newName, Context: context}
	output := &Rename_output{}

	e := self.backendFs.Rename(input, output)

	if e != nil {
                log.Fatalf("Fuse call to backendFs.Rename failed: %v\n", e)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Rename(oldName, newName, context)
	}
	return output.Status
}

func (self *Frontend) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	input := &Mkdir_input{Name: name, Mode: mode, Context: context}
	output := &Mkdir_output{}

	e := self.backendFs.Mkdir(input, output)

	if e != nil {
                log.Fatalf("Fuse call to backendFs.Mkdir failed: %v\n", e)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Mkdir(name, mode, context)
	}
	return output.Status
}

func (self *Frontend) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	input := &Rmdir_input{Name: name, Context: context}
	output := &Rmdir_output{}

	e := self.backendFs.Rmdir(input, output)

	if e != nil {
                log.Fatalf("Fuse call to backendFs.Rmdir failed: %v\n", e)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Rmdir(name, context)
	}
	return output.Status
}

func (self *Frontend) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	input := &Create_input{path, flags, mode, context}
	output := &Create_output{}

	e := self.backendFs.Create(input, output)

	if e != nil {
                log.Fatalf("Fuse call to backendFs.Create failed: %v\n", e)
                e = self.RefreshClient()
                if e != nil { panic(e) }
                return self.Create(path, flags, mode, context)
	}
	fuseFile = &FrontendFile{Backend: self.backendFs, Name: path, Context: context}

	return fuseFile, output.Status
}



// A frontend file is passed to the fuse front end, it has the means to forward the operations to a file on the backed server
type FrontendFile struct {
	Name string
	Backend BackendFs
	Context *fuse.Context
}
func (self *FrontendFile) SetInode(*nodefs.Inode) {} //ok
func (self *FrontendFile) String() string {return fmt.Sprintf("FrontendFile(%v:%v)", self.Backend, self.Name)}
func (self *FrontendFile) InnerFile() nodefs.File {return nil} //ok

func (self *FrontendFile) Read(dest []byte, off int64) (readResult fuse.ReadResult, status fuse.Status) {
	fmt.Println("Read:", self.Name)
	input := &FileRead_input{Path: self.Name, Off: off, BuffLen: len(dest)}
	output := &FileRead_output{Dest: dest, ReadResult: readResult, Status: status}
	e := self.Backend.FileRead(input, output)
	if e != nil {
		// log.Fatalf("backend faild to read file: %v\n", e)
		return nil, fuse.EIO
	}
	return output.ReadResult, output.Status
}

func (self *FrontendFile) Write(data []byte, off int64) (written uint32, code fuse.Status) {
	fmt.Println("Write:", self.Name)
	input := &FileWrite_input{Path: self.Name, Data: data, Off:off, Context: self.Context}
	output := &FileWrite_output{}
	e := self.Backend.FileWrite(input, output)

	if e != nil {
            log.Fatalf("backend faild to write file: %v\n", e)
            return 0, fuse.ENOSYS
            /*
            e = self.RefreshClient()
            if e != nil { panic(e) }
            return self.Write(data, off)
            */
	}

	return output.Written, output.Status
}

func (self *FrontendFile) Flock(flags int) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Flush() fuse.Status {return fuse.ENOSYS} //TODO?

func (self *FrontendFile) Release() {
	/*
	input := &FileRelease_input{self.Name}
	output := &FileRelease_output{}
	self.Backend.FileRelease(input, output)
	*/
}

func (self *FrontendFile) Fsync(flags int) (code fuse.Status) {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Truncate(size uint64) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) GetAttr(out *fuse.Attr) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Chown(uid uint32, gid uint32) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Chmod(perms uint32) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {return fuse.ENOSYS} //TODO?
func (self *FrontendFile) Allocate(off uint64, size uint64, mode uint32) (code fuse.Status) {return fuse.ENOSYS} //TODO?

var _ nodefs.File = new(FrontendFile)
