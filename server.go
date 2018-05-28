package proj

import (
	"fmt"
        "time"
	"encoding/gob"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type ServerFs struct {
	path string
	Addr string
	fs pathfs.FileSystem
	kc *KeeperClient
	openFiles map[string]nodefs.File
	openFlags map[string]uint32
        //coord *ServerCoordinator
}

func NewServerFs(directory, addr string) ServerFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	fs := NewCustomLoopbackFileSystem(directory)
	kc := NewKeeperClient(addr)
	openFiles := make(map[string]nodefs.File)
	openFlags := make(map[string]uint32)

        // initialize keeper, but delay it until after server is served
        go func() {
            time.Sleep(time.Second / 2)
            fmt.Println("init keeper")
            e := kc.Init()
            if e != nil {
                    panic(e)
            }
        }()

        return ServerFs{directory, addr, fs, kc, openFiles, openFlags}
}

func (self *ServerFs) Open(input *Open_input, output *Open_output) error {
	fmt.Println("Open:", input.Name)
	loopbackFile, status := self.fs.Open(input.Name, input.Flags, input.Context)
	if status == fuse.ENOENT {
		fmt.Println("file "+input.Name+" not currently on server")
		kmeta, e := self.kc.Get(input.Name)
		if e != nil {
			panic(e)
		}
		client := NewClientFs(kmeta.Primary.Addr)
		clientFile := &FrontendFile{Name: input.Name, Backend: client, Context: input.Context}
		dest := make([]byte, kmeta.Attr.Size)
		_, readStatus := clientFile.Read(dest, 0)

		newFile, create_status := self.fs.Create(input.Name, input.Flags, kmeta.Attr.Mode, input.Context)
		if create_status != fuse.OK {
			panic(create_status)
		}

		if readStatus == fuse.OK {
			newFile.Write(dest, 0)
		} else { panic("could not read primary copy") }

		fmt.Println("file transferred over")
		loopbackFile = newFile
		status = create_status

		kmeta.Replicas[self.Addr] = ServerFileMeta{self.Addr, 0, 0}
		e = self.kc.Set(input.Name, kmeta)
		if e != nil {
			panic(e)
		}
                // this says we are going to add a replica to this servers
                // metadata
		e = self.kc.AddServer(input.Name, true)
		if e != nil {
			panic(e)
		}
	}

	self.openFiles[input.Name] = loopbackFile
	self.openFlags[input.Name] = input.Flags
	output.Status = status
	return nil
}

func (self *ServerFs) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	// use the keeper to list all the files in the directory
	fmt.Println("opening dir:", input.Name)
	output.Stream, output.Status = self.fs.OpenDir(input.Name, input.Context)
	entries, e := self.kc.GetChildrenAttributes(input.Name)
	if e != nil {
		return e
	}

	output.Stream = entries
	return nil
}

func (self *ServerFs) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	//fmt.Println("GetAttr", input.Name)
	// fetch the attr from zk
	kmeta, e := self.kc.Get(input.Name)
	if e != nil {
		// do nothing
	}
	output.Attr = &kmeta.Attr
	if output.Attr.Ino == 0 {
		output.Attr, output.Status = self.fs.GetAttr(input.Name, input.Context)
	}

	return nil
}

func (self *ServerFs) Rename(input *Rename_input, output *Rename_output) error {
	fmt.Println("Rename:",input.Old,"to",input.New)
	output.Status = self.fs.Rename(input.Old, input.New, input.Context)
	a, _ := self.fs.GetAttr(input.New, input.Context)

	e := self.kc.Create(input.New, *a)
	if e != nil {
		// do nothing for now
		fmt.Println("mv error", e)
		return e
	}

	e = self.kc.Remove(input.Old)
	if e != nil {
		// do nothing for now
		fmt.Println("mv error", e)
		return e
	}
	return nil
}

func (self *ServerFs) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	output.Status = self.fs.Mkdir(input.Name, input.Mode, input.Context)
	// get attributes after make 
	a, _ := self.fs.GetAttr(input.Name, input.Context)

	e := self.kc.Create(input.Name, *a)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerFs) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	output.Status = self.fs.Rmdir(input.Name, input.Context)
	e := self.kc.RemoveDir(input.Name)

	if e != nil {
		return e
	}

	return nil
}

func (self *ServerFs) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	kmeta, e := self.kc.Get(input.Name)
	if e != nil {
		return e
	}
	if self.Addr == kmeta.Primary.Addr {
		e = self.kc.Remove(input.Name)
		if e != nil {
			panic(e)
		}
		status := self.fs.Unlink(input.Name, input.Context)

		for _, replica := range kmeta.Replicas {
			client := NewClientFs(replica.Addr)
			client.Connect()
			e = client.ReplicaUnlink(input, output)
			if e != nil {
				return e
			}
		}

		output.Status = status

	} else {
		client := NewClientFs(kmeta.Primary.Addr)
		client.Connect()
		e = client.Unlink(input, output)
		if e != nil {
			panic(e)
		}
	}

	return nil
}

func (self *ServerFs) ReplicaUnlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("ReplicaUnlink:",input.Name)
	output.Status = self.fs.Unlink(input.Name, input.Context)
	return nil
}

func (self *ServerFs) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)
	loopbackFile, status := self.fs.Create(input.Path, input.Flags, input.Mode, input.Context)
	a, _ := self.fs.GetAttr(input.Path, input.Context)

	// TODO after create check if success or fail
	e := self.kc.Create(input.Path, *a)
	if e != nil {
		return e
	}
	self.openFiles[input.Path] = loopbackFile
	output.Status = status
	return nil
}

func (self *ServerFs) FileRead(input *FileRead_input, output *FileRead_output) error {
	fmt.Println("Read -", "Path:", input.Path)

	output.Dest = make([]byte, input.BuffLen) // recreates the buffer on server for client/server or replaces orignal for local
	output.ReadResult, output.Status = self.openFiles[input.Path].Read(output.Dest, input.Off)

        // increment read count
        e := self.kc.Inc(input.Path, true)
        if e != nil {
            fmt.Errorf("error incrementing read", e)
            return e
        }
	// if output.Status != fuse.OK {
	// 	loopbackFile, _ := self.fs.Open(input.Path, 0, nil)
	// 	self.openFiles[input.Path] = loopbackFile
	// 	output.ReadResult, output.Status = self.openFiles[input.Path].Read(output.Dest, input.Off)
	// }
	fmt.Println("readStatus ", output.Status)
	return nil
}

func (self *ServerFs) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	fmt.Println("Write -", "Path:", input.Path)

	kmeta, e := self.kc.Get(input.Path)
	if e != nil {
		return e
	}
	if self.Addr == kmeta.Primary.Addr {
		fmt.Println("Is the primary, path:",input.Path,"offset:",input.Off)
		written, status := self.openFiles[input.Path].Write(input.Data, input.Off)
		fmt.Println("written:",written,"status:",status)

		for _, replica := range kmeta.Replicas {
			client := NewClientFs(replica.Addr)
			client.Connect()
			e = client.ReplicaFileWrite(input, output)
			if e != nil {
				return e
			}
		}

		output.Written = written
		output.Status = status

		// after writing the file we will go an update the node that we created/modified
		a, _ := self.fs.GetAttr(input.Path, input.Context)

		kmeta.Attr = *a
                // dont call Inc, would be extra call
                kmeta.Primary.WriteCount += 1
		e = self.kc.Set(input.Path, kmeta)
		if e != nil {
			return e
		}

	} else {
		fmt.Println("Not primary, forwarding request to primary")
		client := NewClientFs(kmeta.Primary.Addr)
		client.Connect()
		input.Flags = self.openFlags[input.Path]
		input.Kmeta = kmeta
		e = client.FileWrite(input, output)
		if e != nil {
			return e
		}
	}

	return nil

}

func (self *ServerFs) ReplicaFileWrite(input *FileWrite_input, output *FileWrite_output) error {
	fmt.Println("ReplicaWrite -", "Path:", input.Path)
	output.Written, output.Status = self.openFiles[input.Path].Write(input.Data, input.Off)
	if output.Status != fuse.OK {
	    return fmt.Errorf("failed:",output.Status)
	}
        e := self.kc.Inc(input.Path, false)
        if e != nil {
            return fmt.Errorf("Error incrementing write", e)
        }
	return nil
}

func (self *ServerFs) FileRelease(input *FileRelease_input, output *FileRelease_output) error {
	/*
	fmt.Println("Releasing file", input.Path)
	
	self.openFiles[input.FileId].Release()
	//Removes file from open 
	self.openFiles = append(self.openFiles[:input.FileId], self.openFiles[input.FileId+1:])
	*/
	return nil
}

