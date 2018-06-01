package proj

import (
	// "log"
	"fmt"
	"encoding/gob"
	"github.com/hanwen/go-fuse/fuse"
    "strings"
)

type ServerCoordinator struct {
	path string
	Addr string
	sfsaddr string
	sfs *ServerFS
	kc *KeeperClient
	serverfsm map[string]*ClientFs
	servercoords map[string]*ClientFs
	primary bool
}

func NewServerCoordinator(directory, coordaddr, sfsaddr string) *ServerCoordinator {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	// serve the rpc client
	sfs := NewServerFS(directory, sfsaddr)
	go Serve(sfs)
	// create a new keeper registering the address of this server
	kc := NewKeeperClient(coordaddr, sfsaddr)
	return &ServerCoordinator{path: directory, Addr: coordaddr, sfsaddr: sfsaddr, sfs: sfs, kc: kc}
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
		watch, e := self.kc.AliveWatch()
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
	e := self.sfs.Open(input, output)
	if e != nil {
		// the file doesn't exist on the server
		fmt.Println("file "+input.Name+" not currently on server")
		
		self.checkAndCreatePath(input.Name)

		kmeta, e := self.kc.Get(input.Name)
		if e != nil {
			panic(e)
		}
		client := self.servercoords[kmeta.Primary]
		clientFile := &FrontendFile{Name: input.Name, Backend: client, Context: input.Context}
		dest := make([]byte, kmeta.Attr.Size)
		_, readStatus := clientFile.Read(dest, 0)

		tmpout := Create_output{}
		cin := Create_input{input.Name, input.Flags, kmeta.Attr.Mode, input.Context}
		e = self.sfs.Create(&cin, &tmpout)
		if tmpout.Status != fuse.OK {
			panic(tmpout.Status)
		}
		if readStatus == fuse.OK {
			fi := FileWrite_input{input.Name, dest, 0, input.Context, input.Flags, kmeta}
			fo := FileWrite_output{}
			e = self.sfs.FileWrite(&fi, &fo)
			} else { panic("could not read primary copy") }
			fmt.Println("file transferred over")
			status := tmpout.Status

			kmeta.Replicas = append(kmeta.Replicas, self.sfsaddr)
			e = self.kc.Set(input.Name, kmeta)
			if e != nil {
				panic(e)
			}
			output.Status = status
	}
	return nil
}

func (self *ServerCoordinator) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	// use the keeper to list all the files in the directory
	fmt.Println("opening dir:", input.Name)
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
		self.sfs.GetAttr(input, output)
	}
	return nil
}

func (self *ServerCoordinator) Rename(input *Rename_input, output *Rename_output) error {
	fmt.Println("Rename:",input.Old,"to",input.New)
	
	// get attributes of dir
	kmeta, e := self.kc.Get(input.Old)
	if e != nil {
		panic(e)
	}

	// if this is the primary do renaming
	// else forward task
	if self.Addr == kmeta.Primary {
		// rename in backups
		for _, addr := range kmeta.Replicas {
			fmt.Printf("asking %v to rename file: %v to %v\n", addr, input.Old, input.New)		
			err := self.serverfsm[addr].Rename(input, output)
			if err != nil || output.Status != fuse.OK {
				fmt.Println(input.Old, input.New, ": ", err, output.Status)
				panic(err)
			}
	    }
	    // rename in primary
		err := self.sfs.Rename(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Old, input.New, ": ", err, output.Status)
			panic(err)
		}
		// rename in keeper
		// create the file
		e := self.kc.Create(input.New, kmeta.Attr)
		if e != nil {
			// do nothing for now
			fmt.Println("mv error", e)
			panic(e)
		}
		// Set its metadata
		e = self.kc.Set(input.New, kmeta)
		if e != nil {
			// do nothing for now
			fmt.Println("mv error", e)
			panic(e)
		}
		// remove the old file
		e = self.kc.Remove(input.Old)
		if e != nil {
			// do nothing for now
			fmt.Println("mv error", e)
			panic(e)
		}
	} else {
		err := self.servercoords[kmeta.Primary].Rename(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Old, input.New, ": ", err, output.Status)
			panic(err)
		}		
	}

	return nil
}

func (self *ServerCoordinator) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	fmt.Println("make dir:", input.Name)
		
	self.checkAndCreatePath(input.Name)
	
	self.sfs.Mkdir(input, output)
	e := self.kc.Create(input.Name, *output.Attr)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerCoordinator) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	fmt.Println("remove dir:", input.Name)

	// get attributes of dir
	kmeta, e := self.kc.Get(input.Name)
	if e != nil {
		panic(e)
	}

	// if this is the primary do removing
	// else forward task
	if self.Addr == kmeta.Primary {
		// removing from backups
		for _, addr := range kmeta.Replicas {
			fmt.Printf("asking %v to remove dir: %v\n", addr, input.Name)		
			err := self.serverfsm[addr].Rmdir(input, output)
			if err != nil || output.Status != fuse.OK {
				fmt.Println(input.Name, ": ", err, output.Status)
				panic(err)
			}
	    }
	    // remove from primary
		err := self.sfs.Rmdir(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Name, ": ", err, output.Status)
			panic(err)
		}
		// remove from keeper
		e = self.kc.RemoveDir(input.Name)
		if e != nil {
			return e
		}
	} else {
		err := self.servercoords[kmeta.Primary].Rmdir(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Name, ": ", err, output.Status)
			panic(err)
		}		
	}


	return nil
}

func (self *ServerCoordinator) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	kmeta, e := self.kc.Get(input.Name)
	if e != nil { return e }
	if self.Addr == kmeta.Primary {
		e = self.kc.Remove(input.Name)
		if e != nil {
			panic(e)
		}
		self.sfs.Unlink(input, output)
		for _, replicaAddr := range kmeta.Replicas {
			// get client
			client := self.serverfsm[replicaAddr]
			e = client.Unlink(input, output)
			if e != nil {
				return e
			}
		}
	} else {
		client := self.servercoords[kmeta.Primary]
		e = client.Unlink(input, output)
		if e != nil {
			panic(e)
		}
	}

	return nil
}

func (self *ServerCoordinator) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)
		
	self.checkAndCreatePath(input.Path)
	
	self.sfs.Create(input, output)

	e := self.kc.Create(input.Path, *output.Attr)
	if e != nil {
		return e
	}
	return nil
}

func (self *ServerCoordinator) FileRead(input *FileRead_input, output *FileRead_output) error {
	fmt.Println("Read -", "Path:", input.Path)
	self.sfs.FileRead(input, output)
	return nil
}

func (self *ServerCoordinator) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	fmt.Println("Write -", "Path:", input.Path)

	kmeta, e := self.kc.Get(input.Path)
	if e != nil {
		return e
	}
	if self.Addr == kmeta.Primary {
		fmt.Println("Is the primary, path:",input.Path,"offset:",input.Off)
		self.sfs.FileWrite(input, output)

		for _, replicaAddr := range kmeta.Replicas {
			fmt.Println("clients", self.serverfsm, replicaAddr)
			client := self.serverfsm[replicaAddr]
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
		client := self.servercoords[kmeta.Primary]
		input.Kmeta = kmeta
		e = client.FileWrite(input, output)
		if e != nil {
			return e
		}
	}

	return nil

}

func (self *ServerCoordinator) FileRelease(input *FileRelease_input, output *FileRelease_output) error {
	fmt.Println("Release -", "Path:", input.Path)
	/*
	fmt.Println("Releasing file", input.Path)
	
	self.openFiles[input.FileId].Release()
	//Removes file from open 
	self.openFiles = append(self.openFiles[:input.FileId], self.openFiles[input.FileId+1:])
	*/
	return nil
}

/*
this function will split the path and check that all directories exits
if a directory does not exist it will create it on this coordinators ServerFs

it does not do the last element of the split because it is the thing the calling 
op is woking on
*/
func (self *ServerCoordinator) checkAndCreatePath(path string) {
	dirs := strings.Split(path, "/")
	// remove all but last filename or dir
	if len(dirs[len(dirs)-1]) == 0 {
		dirs = dirs[:len(dirs)-2]
	} else {
		dirs = dirs[:len(dirs)-1]
	}

	// for each directory in the path
	for index, _ := range dirs {
		curPath := strings.Join(dirs[:index+1], "/")
		kmeta, e := self.kc.Get(curPath)
		if e != nil {
			panic(e)
		}

		// check if it alread has it
		if self.Addr == kmeta.Primary {
			continue
		}
		for _, addr := range kmeta.Replicas {
        	if self.sfsaddr == addr {
            	continue
        	}
        }
		
		fmt.Println("directory "+curPath+" not currently on server")

        // create the dir lovally
        /* 
        TODO: HOW ARE WE HANDLING LOCAL CHANGES? 
        GOING THROUGH the ServerFS will cause loopback rpc
        */
        input := &Mkdir_input{Name: curPath, Mode: kmeta.Attr.Mode}
        output := &Mkdir_output{}

		err := self.sfs.Mkdir(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(curPath, ": ", err, output.Status)
			panic(err)
		}

        // update keeper
		kmeta.Replicas = append(kmeta.Replicas, self.sfsaddr)
		e = self.kc.Set(curPath, kmeta)
		if e != nil {
			panic(e)
		}

	}
}

// assert that ServerCoordinator implements BackendFs
var _ BackendFs = new(ServerCoordinator)
