package proj

import (
	"fmt"
	"log"
	"strings"
	"encoding/gob"
	"net"
	"net/http"
	"net/rpc"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type ServerFS struct {
	path string
	Addr string
	fs pathfs.FileSystem
	openFiles map[string]nodefs.File
	openFlags map[string]uint32
}

func NewServerFS(directory, addr string) *ServerFS {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	fs := NewCustomLoopbackFileSystem(directory)
	openFiles := make(map[string]nodefs.File)
	openFlags := make(map[string]uint32)

	return &ServerFS{directory, addr, fs, openFiles, openFlags}
}

func Serve(sfs *ServerFS) {
    // setup rpc server
	port := strings.Split(sfs.Addr, ":")[1]
	server := rpc.NewServer()
	e := server.RegisterName("BackendFs", sfs)
	l, e := net.Listen ("tcp",":"+port)
	if e != nil {
		log.Fatal(e)
	}

    // serve
	log.Printf("key-value store serving directory \"%s\" on %s", sfs.path, sfs.Addr)
	e = http.Serve(l, server)
	if e != nil {
		log.Fatal(e)
	}
}

func (self *ServerFS) Open(input *Open_input, output *Open_output) error {
	loopbackFile, status := self.fs.Open(input.Name, input.Flags, input.Context)
	if status != fuse.ENOENT {
		self.openFiles[input.Name] = loopbackFile
		self.openFlags[input.Name] = input.Flags
		output.Status = status
		return nil
	}
	return fmt.Errorf("Error: File Does Not Exist")
}

func (self *ServerFS) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	output.Stream, output.Status = self.fs.OpenDir(input.Name, input.Context)
	return nil
}

func (self *ServerFS) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	output.Attr, output.Status = self.fs.GetAttr(input.Name, input.Context)
	return nil
}

func (self *ServerFS) Rename(input *Rename_input, output *Rename_output) error {
	output.Status = self.fs.Rename(input.Old, input.New, input.Context)
	output.Attr, _ = self.fs.GetAttr(input.New, input.Context)
	return nil
}

func (self *ServerFS) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	fmt.Println("making dir", input.Name)
	e := self.checkAndCreatePath(input.Name, input.Context)
	if e != nil { panic(e) }
	output.Status = self.fs.Mkdir(input.Name, input.Mode, input.Context)
	// get attributes after make 
	output.Attr, _ = self.fs.GetAttr(input.Name, input.Context)
	return nil
}

func (self *ServerFS) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	output.Status = self.fs.Rmdir(input.Name, input.Context)
	return nil
}

func (self *ServerFS) Unlink(input *Unlink_input, output *Unlink_output) error {
	output.Status = self.fs.Unlink(input.Name, input.Context)
	// close the file if it was open
	delete(self.openFiles, input.Name)
	delete(self.openFlags, input.Name)
	return nil
}

func (self *ServerFS) Create(input *Create_input, output *Create_output) error {
	//delete(self.openFiles, input.Path)
	//delete(self.openFlags, input.Path)
	// before creating the file we need to ensure the path is good to go
	e := self.checkAndCreatePath(input.Path, input.Context)
	if e != nil { panic(e) }
	loopbackFile, status := self.fs.Create(input.Path, input.Flags, input.Mode, input.Context)
	output.Attr, _ = self.fs.GetAttr(input.Path, input.Context)
	self.openFiles[input.Path] = loopbackFile
	output.Status = status
	return nil
}

func (self *ServerFS) FileRead(input *FileRead_input, output *FileRead_output) error {
	output.Dest = make([]byte, input.BuffLen) // recreates the buffer on server for client/server or replaces orignal for local
	output.ReadResult, output.Status = self.openFiles[input.Path].Read(output.Dest, input.Off)
	// if output.Status != fuse.OK {
	// 	loopbackFile, _ := self.fs.Open(input.Path, 0, nil)
	// 	self.openFiles[input.Path] = loopbackFile
	// 	output.ReadResult, output.Status = self.openFiles[input.Path].Read(output.Dest, input.Off)
	// }
	return nil
}

func (self *ServerFS) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	if _, ok := self.openFiles[input.Path]; !ok {
		fi := Create_input{input.Path, input.Flags, 0755, input.Context}
		fo := Create_output{}
		e := self.Create(&fi, &fo)
		if e != nil { panic(e) }
	}
	output.Written, output.Status = self.openFiles[input.Path].Write(input.Data, input.Off)
	output.Attr, _ = self.fs.GetAttr(input.Path, input.Context)
	return nil
}

func (self *ServerFS) FileRelease(input *FileRelease_input, output *FileRelease_output) error {
	/*
	fmt.Println("Releasing file", input.Path)
	
	self.openFiles[input.FileId].Release()
	//Removes file from open 
	self.openFiles = append(self.openFiles[:input.FileId], self.openFiles[input.FileId+1:])
	*/
	return nil
}

func (self *ServerFS) checkAndCreatePath(path string, context *fuse.Context) error {
	dirs := strings.Split(path, "/")
	// remove all but last filename or dir
	if len(dirs[len(dirs)-1]) == 0 {
		dirs = dirs[:len(dirs)-2]
	} else {
		dirs = dirs[:len(dirs)-1]
	}

	for index, _ := range dirs {
		curPath := strings.Join(dirs[:index+1], "/")
		outStatus := self.fs.Mkdir(curPath, GeoFSMode, context)
		if outStatus != fuse.OK && outStatus != 17 {
			// status 17 is file already exists, couldnt find const
			fmt.Println(curPath, ": ", outStatus)
			return fmt.Errorf("Error Creating Path: %v %v\n", path, outStatus)
		}
	}
	return nil
}
