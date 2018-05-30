package proj

import (
	// "log"
	"fmt"
	"encoding/gob"
	"github.com/hanwen/go-fuse/fuse"
)

type ServerCoordinator struct {
	path string
	Addr string
        fsaddr string
	fs *ServerFS
	kc *KeeperClient
	serverfsm map[string]*ClientFs
	servercoords map[string]*ClientFs
        primary bool
}

func NewServerCoordinator(directory, coordaddr, fsaddr string) *ServerCoordinator {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
        // serve the rpc client
        fs := NewServerFS(directory, fsaddr)
        go Serve(fs)
        // create a new keeper registering the address of this server
	kc := NewKeeperClient(coordaddr, fsaddr)
        return &ServerCoordinator{path: directory, Addr: coordaddr, fsaddr: fsaddr, fs: fs, kc: kc}
}

func (self *ServerCoordinator) Init() error {
        fmt.Println("init keeper")
        e := self.kc.Init()
        if e != nil {
                return e
        }
        self.servercoords, self.serverfsm, e = self.kc.GetBackendMaps()
        if e != nil { return e }
        self.Watch()
        return nil
}

func (self *ServerCoordinator) Watch() error {
    for {
        // might want to use backs
        _, watch, e := self.kc.AliveWatch()
        if e != nil { return e }
        // something changed so go update backend maps
        servercoords, serverfsm, e := self.kc.GetBackendMaps()
        if e != nil { return e }
        // for now immediately update
        self.servercoords = servercoords
        self.serverfsm = serverfsm
        <-watch
    }
}

func (self *ServerCoordinator) Open(input *Open_input, output *Open_output) error {
	fmt.Println("Open:", input.Name)
	e := self.fs.Open(input, output)
	if e != nil {
            // the file doesn't exist on the server
            fmt.Println("file "+input.Name+" not currently on server")
            kmeta, e := self.kc.Get(input.Name)
            if e != nil {
                    panic(e)
            }
            client := self.servercoords[kmeta.Primary.Addr]
            clientFile := &FrontendFile{Name: input.Name, Backend: client, Context: input.Context}
            dest := make([]byte, kmeta.Attr.Size)
            _, readStatus := clientFile.Read(dest, 0)

            // TODO directory stuff here
            tmpout := Create_output{}
            cin := Create_input{input.Name, input.Flags, kmeta.Attr.Mode, input.Context}
            e = self.fs.Create(&cin, &tmpout)
            if tmpout.Status != fuse.OK {
                    panic(tmpout.Status)
            }
            if readStatus == fuse.OK {
                fi := FileWrite_input{input.Name, dest, 0, input.Context, input.Flags, kmeta}
                fo := FileWrite_output{}
                fmt.Println("about to write file")
                e = self.fs.FileWrite(&fi, &fo)
            } else { panic("could not read primary copy") }
            fmt.Println("file transferred over")
            status := tmpout.Status

            kmeta.Replicas[self.fsaddr] = ServerFileMeta{self.fsaddr, 0, 0}
            e = self.kc.Set(input.Name, kmeta)
            if e != nil {
                    panic(e)
            }
            e = self.kc.AddServerMeta(input.Name, self.fsaddr, true)
            output.Status = status
	}
	return nil
}

func (self *ServerCoordinator) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	// use the keeper to list all the files in the directory
	fmt.Println("opening dir:", input.Name)
        self.fs.OpenDir(input, output)
	entries, e := self.kc.GetChildrenAttributes(input.Name)
	if e != nil {
		return e
	}

	output.Stream = entries
	return nil
}

func (self *ServerCoordinator) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
        fmt.Println("get attr:", input.Name)
	// fetch the attr from zk
	kmeta, e := self.kc.Get(input.Name)
	if e != nil {
		// do nothing
	}
	output.Attr = &kmeta.Attr
	if output.Attr.Ino == 0 {
                self.fs.GetAttr(input, output)
	}
	return nil
}

func (self *ServerCoordinator) Rename(input *Rename_input, output *Rename_output) error {
	fmt.Println("Rename:",input.Old,"to",input.New)
	self.fs.Rename(input, output)

        // create adds relevant metadata
	e := self.kc.Create(input.New, *output.Attr)
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
        self.kc.RemoveServerMeta(input.Old, self.Addr, false)
	return nil
}

func (self *ServerCoordinator) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	self.fs.Mkdir(input, output)
	e := self.kc.Create(input.Name, *output.Attr)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerCoordinator) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	self.fs.Rmdir(input, output)
	e := self.kc.RemoveDir(input.Name)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerCoordinator) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	kmeta, e := self.kc.Get(input.Name)
	if e != nil { return e }
	if self.Addr == kmeta.Primary.Addr {
		e = self.kc.Remove(input.Name)
		if e != nil {
			panic(e)
		}
		self.fs.Unlink(input, output)
                self.kc.RemoveServerMeta(input.Name, self.Addr, false)
		for _, replica := range kmeta.Replicas {
                    // get client
                    client := self.serverfsm[replica.Addr]
                    e = client.Unlink(input, output)
                    if e != nil {
                            return e
                    }
                    self.kc.RemoveServerMeta(input.Name, replica.Addr, true)
		}
	} else {
                client := self.servercoords[kmeta.Primary.Addr]
		e = client.Unlink(input, output)
		if e != nil {
			panic(e)
		}
	}

	return nil
}

func (self *ServerCoordinator) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)
	self.fs.Create(input, output)

	e := self.kc.Create(input.Path, *output.Attr)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerCoordinator) FileRead(input *FileRead_input, output *FileRead_output) error {
        self.fs.FileRead(input, output)
	return nil
}

func (self *ServerCoordinator) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	fmt.Println("Write -", "Path:", input.Path)

	kmeta, e := self.kc.Get(input.Path)
	if e != nil {
		return e
	}
	if self.Addr == kmeta.Primary.Addr {
		fmt.Println("Is the primary, path:",input.Path,"offset:",input.Off)
                self.fs.FileWrite(input, output)

		for _, replica:= range kmeta.Replicas {
			client := self.serverfsm[replica.Addr]
			e = client.FileWrite(input, output)
			if e != nil {
				return e
			}
		}

		kmeta.Attr = *output.Attr
		e = self.kc.Set(input.Path, kmeta)
		if e != nil {
			return e
		}

	} else {
		fmt.Println("Not primary, forwarding request to primary coordinator")
                client := self.servercoords[kmeta.Primary.Addr]
		input.Kmeta = kmeta
		e = client.FileWrite(input, output)
		if e != nil {
			return e
		}
	}

	return nil

}

func (self *ServerCoordinator) FileRelease(input *FileRelease_input, output *FileRelease_output) error {
	/*
	fmt.Println("Releasing file", input.Path)
	
	self.openFiles[input.FileId].Release()
	//Removes file from open 
	self.openFiles = append(self.openFiles[:input.FileId], self.openFiles[input.FileId+1:])
	*/
	return nil
}

// assert that ServerCoordinator implements BackendFs
var _ BackendFs = new(ServerCoordinator)
