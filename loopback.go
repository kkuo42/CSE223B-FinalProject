package proj

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"sync"
	"fmt"
	"time"
	"unsafe"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

///////////////////////////////////////////////////////////////////////////////
///////////////             loopback file system               ////////////////
///////////////////////////////////////////////////////////////////////////////
// https://github.com/hanwen/go-fuse/blob/master/fuse/pathfs/loopback.go

type CustomLoopbackFileSystem struct {
	// TODO - this should need default fill in.
	pathfs.FileSystem
	Root string
}

// A FUSE filesystem that shunts all request to an underlying file
// system.  Its main purpose is to provide test coverage without
// having to build a synthetic filesystem.
func NewCustomLoopbackFileSystem(root string) pathfs.FileSystem {
	// Make sure the Root path is absolute to avoid problems when the
	// application changes working directory.
	root, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	return &CustomLoopbackFileSystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
		Root:       root,
	}
}

func (fs *CustomLoopbackFileSystem) StatFs(name string) *fuse.StatfsOut {
	s := syscall.Statfs_t{}
	err := syscall.Statfs(fs.GetPath(name), &s)
	if err == nil {
		out := &fuse.StatfsOut{}
		out.FromStatfsT(&s)
		return out
	}
	return nil
}

func (fs *CustomLoopbackFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
}

func (fs *CustomLoopbackFileSystem) OnUnmount() {}

func (fs *CustomLoopbackFileSystem) GetPath(relPath string) string {
	return filepath.Join(fs.Root, relPath)
}

func (fs *CustomLoopbackFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	fullPath := fs.GetPath(name)
	var err error = nil
	st := syscall.Stat_t{}
	if name == "" {
		// When GetAttr is called for the toplevel directory, we always want
		// to look through symlinks.
		err = syscall.Stat(fullPath, &st)
	} else {
		err = syscall.Lstat(fullPath, &st)
	}

	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	a = &fuse.Attr{}
	a.FromStat(&st)
	return a, fuse.OK
}

func (fs *CustomLoopbackFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	// What other ways beyond O_RDONLY are there to open
	// directories?
	f, err := os.Open(fs.GetPath(name))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	want := 500
	output := make([]fuse.DirEntry, 0, want)
	for {
		infos, err := f.Readdir(want)
		for i := range infos {
			// workaround for https://code.google.com/p/go/issues/detail?id=5960
			if infos[i] == nil {
				continue
			}
			n := infos[i].Name()
			d := fuse.DirEntry{
				Name: n,
			}
			if s := fuse.ToStatT(infos[i]); s != nil {
				d.Mode = uint32(s.Mode)
				d.Ino = s.Ino
			} else {
				log.Printf("ReadDir entry %q for %q has no stat info", n, name)
			}
			output = append(output, d)
		}
		if len(infos) < want || err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Readdir() returned err:", err)
			break
		}
	}
	f.Close()

	return output, fuse.OK
}


/* this function was rewritten to pass the custom file */
func (fs *CustomLoopbackFileSystem) Open(name string, flags uint32, context *fuse.Context) (fuseFile nodefs.File, status fuse.Status) {
	f, err := os.OpenFile(fs.GetPath(name), int(flags), 0)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return NewCustomLoopbackFile(f), fuse.OK
}

func (fs *CustomLoopbackFileSystem) Chmod(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	err := os.Chmod(fs.GetPath(path), os.FileMode(mode))
	return fuse.ToStatus(err)
}

func (fs *CustomLoopbackFileSystem) Chown(path string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Chown(fs.GetPath(path), int(uid), int(gid)))
}

func (fs *CustomLoopbackFileSystem) Truncate(path string, offset uint64, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Truncate(fs.GetPath(path), int64(offset)))
}

func (fs *CustomLoopbackFileSystem) Readlink(name string, context *fuse.Context) (out string, code fuse.Status) {
	f, err := os.Readlink(fs.GetPath(name))
	return f, fuse.ToStatus(err)
}

func (fs *CustomLoopbackFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Mknod(fs.GetPath(name), mode, int(dev)))
}

func (fs *CustomLoopbackFileSystem) Mkdir(path string, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Mkdir(fs.GetPath(path), os.FileMode(mode)))
}

// Don't use os.Remove, it removes twice (unlink followed by rmdir).
func (fs *CustomLoopbackFileSystem) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Unlink(fs.GetPath(name)))
}

func (fs *CustomLoopbackFileSystem) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Rmdir(fs.GetPath(name)))
}

func (fs *CustomLoopbackFileSystem) Symlink(pointedTo string, linkName string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Symlink(pointedTo, fs.GetPath(linkName)))
}

func (fs *CustomLoopbackFileSystem) Rename(oldPath string, newPath string, context *fuse.Context) (codee fuse.Status) {
	err := os.Rename(fs.GetPath(oldPath), fs.GetPath(newPath))
	return fuse.ToStatus(err)
}

func (fs *CustomLoopbackFileSystem) Link(orig string, newName string, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(os.Link(fs.GetPath(orig), fs.GetPath(newName)))
}

func (fs *CustomLoopbackFileSystem) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ToStatus(syscall.Access(fs.GetPath(name), mode))
}

func (fs *CustomLoopbackFileSystem) Create(path string, flags uint32, mode uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	f, err := os.OpenFile(fs.GetPath(path), int(flags)|os.O_CREATE, os.FileMode(mode))
	return NewCustomLoopbackFile(f), fuse.ToStatus(err)
}



///////////////////////////////////////////////////////////////////////////////
///////////////               loopback file                    ////////////////
///////////////////////////////////////////////////////////////////////////////
//https://github.com/hanwen/go-fuse/blob/master/fuse/nodefs/files.go

func NewCustomLoopbackFile(f *os.File) nodefs.File {
	return &CustomLoopbackFile{File: f}
}

type CustomLoopbackFile struct {
	File *os.File

	// os.File is not threadsafe. Although fd themselves are
	// constant during the lifetime of an open file, the OS may
	// reuse the fd number after it is closed. When open races
	// with another close, they may lead to confusion as which
	// file gets written in the end.
	lock sync.Mutex
}

func (f *CustomLoopbackFile) InnerFile() nodefs.File {
	return nil
}

func (f *CustomLoopbackFile) SetInode(n *nodefs.Inode) {
}

func (f *CustomLoopbackFile) String() string {
	return fmt.Sprintf("CustomLoopbackFile(%s)", f.File.Name())
}

func (f *CustomLoopbackFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	f.lock.Lock()
	// This is not racy by virtue of the kernel properly
	// synchronizing the open/write/close.

	// r := fuse.ReadResultFd(f.File.Fd(), off, len(buf))

	/* 
	replacing the call to ReadResultFd with ReadResultData code so the buffer
	is passed via RPC 
	*/

	sz := len(buf)
	if len(buf) < sz {
		sz = len(buf)
	}

	n, err := syscall.Pread(int(f.File.Fd()), buf[:sz], off)
	if err == io.EOF {
		err = nil
	}

	if n < 0 {
		n = 0
	}

	r := NewCustomReadResultData(buf[:n])

	f.lock.Unlock()
	return r, fuse.ToStatus(err)
}

func (f *CustomLoopbackFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	f.lock.Lock()
	n, err := f.File.WriteAt(data, off)
	f.lock.Unlock()
	return uint32(n), fuse.ToStatus(err)
}

func (f *CustomLoopbackFile) Release() {
	f.lock.Lock()
	f.File.Close()
	f.lock.Unlock()
}

func (f *CustomLoopbackFile) Flush() fuse.Status {
	f.lock.Lock()

	// Since Flush() may be called for each dup'd fd, we don't
	// want to really close the file, we just want to flush. This
	// is achieved by closing a dup'd fd.
	newFd, err := syscall.Dup(int(f.File.Fd()))
	f.lock.Unlock()

	if err != nil {
		return fuse.ToStatus(err)
	}
	err = syscall.Close(newFd)
	return fuse.ToStatus(err)
}

func (f *CustomLoopbackFile) Fsync(flags int) (code fuse.Status) {
	f.lock.Lock()
	r := fuse.ToStatus(syscall.Fsync(int(f.File.Fd())))
	f.lock.Unlock()

	return r
}

func (f *CustomLoopbackFile) Flock(flags int) fuse.Status {
	f.lock.Lock()
	r := fuse.ToStatus(syscall.Flock(int(f.File.Fd()), flags))
	f.lock.Unlock()

	return r
}

func (f *CustomLoopbackFile) Truncate(size uint64) fuse.Status {
	f.lock.Lock()
	r := fuse.ToStatus(syscall.Ftruncate(int(f.File.Fd()), int64(size)))
	f.lock.Unlock()

	return r
}

func (f *CustomLoopbackFile) Chmod(mode uint32) fuse.Status {
	f.lock.Lock()
	r := fuse.ToStatus(f.File.Chmod(os.FileMode(mode)))
	f.lock.Unlock()

	return r
}

func (f *CustomLoopbackFile) Chown(uid uint32, gid uint32) fuse.Status {
	f.lock.Lock()
	r := fuse.ToStatus(f.File.Chown(int(uid), int(gid)))
	f.lock.Unlock()

	return r
}

func (f *CustomLoopbackFile) GetAttr(a *fuse.Attr) fuse.Status {
	st := syscall.Stat_t{}
	f.lock.Lock()
	err := syscall.Fstat(int(f.File.Fd()), &st)
	f.lock.Unlock()
	if err != nil {
		return fuse.ToStatus(err)
	}
	a.FromStat(&st)

	return fuse.OK
}

func (f *CustomLoopbackFile) Allocate(off uint64, sz uint64, mode uint32) fuse.Status {
	f.lock.Lock()
	err := syscall.Fallocate(int(f.File.Fd()), mode, int64(off), int64(sz))
	f.lock.Unlock()
	if err != nil {
		return fuse.ToStatus(err)
	}
	return fuse.OK
}

func futimens(fd int, times *[2]syscall.Timespec) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_UTIMENSAT, uintptr(fd), 0, uintptr(unsafe.Pointer(times)), uintptr(0), 0, 0)
	if e1 != 0 {
		err = syscall.Errno(e1)
	}
	return
}
func (f *CustomLoopbackFile) Utimens(a *time.Time, m *time.Time) fuse.Status {
	var ts [2]syscall.Timespec
	ts[0] = fuse.UtimeToTimespec(a)
	ts[1] = fuse.UtimeToTimespec(m)
	f.lock.Lock()
	err := futimens(int(f.File.Fd()), &ts)
	f.lock.Unlock()
	return fuse.ToStatus(err)
}



///////////////////////////////////////////////////////////////////////////////
///////////////               ReadResult                       ////////////////
///////////////////////////////////////////////////////////////////////////////
// https://github.com/hanwen/go-fuse/blob/master/fuse/read.go
/* 
this was needed to because it was not exported and therefore could be rpc 
registered 
*/
// ReadResultData is the read return for returning bytes directly.
type CustomReadResultData struct {
	// Raw bytes for the read.
	Data []byte
}

func (r *CustomReadResultData) Size() int {
	return len(r.Data)
}

func (r *CustomReadResultData) Done() {
}

func (r *CustomReadResultData) Bytes(buf []byte) ([]byte, fuse.Status) {
	return r.Data, fuse.OK
}

func NewCustomReadResultData(b []byte) fuse.ReadResult {
	return &CustomReadResultData{b}
}
