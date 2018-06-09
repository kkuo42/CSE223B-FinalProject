package proj

import (
	// "log"
	"fmt"
	"math"
	"time"
	"strings"
	"strconv"
	"encoding/gob"
	"encoding/json"
	"github.com/hanwen/go-fuse/fuse"
)

type ServerCoordinator struct {
	Path string
	Addr string
	SFSAddr string
	sfs *ServerFS
	kc *KeeperClient
	serverfsm map[string]*ClientFs
	servercoords map[string]*ClientFs
	fileLocks map[string]chan int
	runningbalance bool
}

func NewServerCoordinator(directory, coordaddr, sfsaddr string) *ServerCoordinator {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	// serve the rpc client
	sfs := NewServerFS(directory, sfsaddr, coordaddr)
	go Serve(sfs)
	// create a new keeper registering the address of this server
	kc := NewKeeperClient(coordaddr, sfsaddr)
	return &ServerCoordinator{Path: directory, Addr: coordaddr, SFSAddr: sfsaddr, sfs: sfs, kc: kc}
}

func (self *ServerCoordinator) Init() error {
	fmt.Println("init keeper")
	e := self.kc.Init()
	if e != nil {
		return e
	}
	self.fileLocks = make(map[string]chan int)
	self.runningbalance = false
	//self.servercoords, self.serverfsm, e = self.kc.GetBackendMaps()
	if e != nil { return e }
	go self.Watch()
	return nil
}

func getCoordLeader(backs []string) (minback string, e error) {
	min := math.MaxInt64
	for _, back := range backs {
		split := strings.Split(back, "_")
		seqnum, e := strconv.Atoi(split[1])
		if e != nil {
			return "", fmt.Errorf("Error converting string to int in getCoordLeader %v\n", e)
		}
		if seqnum < min {
			minback = split[0]
			min = seqnum
		}
	}
	return minback, nil
}

// TODO come up with multiple balance functions...
// randomly distributed
// all
// depending when last balanced
func (self *ServerCoordinator) balance() error {
	// can edit time in config.go
	ticker := time.NewTicker(BalanceTime)
	for {
		select {
			case <- ticker.C:
				fmt.Println("BALANCE HERE")
				// we will go and get all files in the directory tree
				// probably easiest to query zk for all of the nodes primary
				// metadata
				alivemeta, e := self.kc.Children("/alivemeta")
				if e != nil { return e }
				for _, addr := range alivemeta {
					var sm ServerMeta
					metadata, e := self.kc.Get("/alivemeta/"+addr)
					if e != nil { return e }
					e = json.Unmarshal(metadata, &sm)
					for _, path := range sm.PrimaryFor {
						// swap for everything
						self.SwapPathPrimary(path, false)
					}
				}
		}
	}
}

func (self *ServerCoordinator) balanceFail(alivemeta []string) error {
	// find the nodes that no longer exist
	deadnodes := []string{}
	for _, addr := range alivemeta {
		_, okc := self.servercoords[addr]
		_, okf := self.serverfsm[addr]
		if !okc && !okf {
			// if coordm and fsm dont have addr, then dead or just joined
			deadnodes = append(deadnodes, addr)
		}
	}
	for _, addr := range deadnodes {
		// get the metadata and then add files to file arrays
		var sm ServerMeta
		metadata, e := self.kc.Get("/alivemeta/"+addr)
		if e != nil { return e }
		e = json.Unmarshal(metadata, &sm)
		// go and do the respective operation to move metadata
		for _, path := range sm.PrimaryFor {
			//primaryfiles = append(primaryfiles, path)
			self.SwapPathPrimary(path, true)
		}
		for _, path := range sm.ReplicaFor {
			self.RemovePathBackup(path, addr, true)
		}
		// remove dead node from alivemeta
		self.kc.Delete("/alivemeta/"+addr)
	}

	return nil
}

func (self *ServerCoordinator) Watch() error {
	for {
		// prep if something is wrong
		backs, watch, e := self.kc.AliveWatch()
		if e != nil { return e }
		// need to get maps before getting meta, so maps are never
		// more up to date than alivemeta in the case of concurrent join
		self.servercoords, self.serverfsm, e = self.kc.GetBackendMaps()
		if e != nil { return e }
		oldalive, e := self.kc.Children("/alivemeta")
		if e != nil { return e }

		// set new coordinator
		coordlead, e := getCoordLeader(backs)
		if e != nil { return e }
		fmt.Println("coordleader is:", coordlead)

		// something has changed so get new maps
		if e != nil { return e }

		if coordlead == self.Addr {
			// something has changed so we will attempt to rebalance if necessary
			self.balanceFail(oldalive)
			if !self.runningbalance {
				self.runningbalance = true
				go self.balance()
			}
		}

		<-watch
	}
}

func (self *ServerCoordinator) Open(input *Open_input, output *Open_output) error {
	fmt.Println("Open:", input.Name)
	e := self.sfs.Open(input, output)

	// TODO: Maybe call keeper .get first because what if the file is lingering? is release/unlink correct?

	if e != nil {
		// the file doesn't exist on the server
		fmt.Println("file "+input.Name+" not currently on server")

		kmeta, e := self.kc.GetData(input.Name)
		if e != nil {
			panic(e)
		}

		fmt.Println("kmeta coordaddr", kmeta.Primary.CoordAddr)
		client := self.servercoords[kmeta.Primary.CoordAddr]
		clientFile := &FrontendFile{Name: input.Name, Backend: client, Context: input.Context, Addr: self.SFSAddr}

		dest := make([]byte, kmeta.Attr.Size)
		_, readStatus := clientFile.Read(dest, 0)

		tmpout := Create_output{}
		cin := Create_input{input.Name, input.Flags, kmeta.Attr.Mode, input.Context}
		e = self.sfs.Create(&cin, &tmpout)
		if tmpout.Status != fuse.OK {
			panic(tmpout.Status)
		}
		if readStatus == fuse.OK {
			fi := FileWrite_input{input.Name, dest, 0, input.Context, input.Flags, kmeta, self.SFSAddr}
			fo := FileWrite_output{}
			e = self.sfs.FileWrite(&fi, &fo)
			} else { panic("could not read primary copy") }
			fmt.Println("file transferred over")
			status := tmpout.Status

			kmeta.Replicas[self.SFSAddr] = ServerFileMeta{self.Addr, self.SFSAddr, 0, 0}
			e = self.kc.Set(input.Name, kmeta)
			if e != nil {
				panic(e)
			}
		    e = self.kc.AddServerMeta(input.Name, self.SFSAddr, true)

			output.Status = status
			//self.kc.Inc(input.Name, true)
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
	kmeta, e := self.kc.GetData(input.Name)
	if e != nil {
		// do nothing
		if e.Error() == "Deleted boolean" {
			self.sfs.GetAttr(input, output)
			return nil
		}
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
	kmeta, e := self.kc.GetData(input.Old)
	if e != nil {
		panic(e)
	}

	// if this is the primary do renaming
	// else forward task
	if self.Addr == kmeta.Primary.CoordAddr {

		Lock(self, input.Old)
		Lock(self, input.New)

		// rename in backups
		for _, replica:= range kmeta.Replicas {
			fmt.Printf("asking %v to rename file: %v to %v\n", replica.SFSAddr, input.Old, input.New)
			err := self.serverfsm[replica.SFSAddr].Rename(input, output)
			if err != nil || output.Status != fuse.OK {
				fmt.Println(input.Old, input.New, ": ", err, output.Status)
				panic(err)
			}
			e := self.kc.AddServerMeta(input.New, replica.SFSAddr, true)
			if e != nil { return e }
		}
	        // rename in primary
		err := self.sfs.Rename(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Old, input.New, ": ", err, output.Status)
			panic(err)
		}
		// rename in keeper
		// create the file
		_, e := self.kc.Create(input.New, kmeta.Attr, kmeta.Deleted)
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
		e = self.kc.Remove(input.Old, kmeta)
		if e != nil {
			// do nothing for now
			fmt.Println("mv error", e)
			panic(e)
		}

		Unlock(self, input.Old)
		Unlock(self, input.New)

	} else {
		err := self.servercoords[kmeta.Primary.CoordAddr].Rename(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Old, input.New, ": ", err, output.Status)
			panic(err)
		}
	}
	return nil
}

func (self *ServerCoordinator) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	self.sfs.Mkdir(input, output)
	replica, e := self.kc.Create(input.Name, *output.Attr, false)
	if e != nil {
		return e
	}
	if replica != "" {
		// create the dir on the replica as well
		e = self.serverfsm[replica].Mkdir(input, output)
		if e != nil {
			return e
		}
	}
	return nil
}

func (self *ServerCoordinator) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	fmt.Println("remove dir:", input.Name)

	// get attributes of dir
	kmeta, e := self.kc.GetData(input.Name)
	if e != nil {
		panic(e)
	}
	// TODO probably unecessary since we arent keeping track

	// if this is the primary do removing
	// else forward task
	if self.Addr == kmeta.Primary.CoordAddr {
		// removing from backups
		for _, replica := range kmeta.Replicas {
			fmt.Printf("asking %v to remove dir: %v\n", replica.SFSAddr, input.Name)
			err := self.serverfsm[replica.SFSAddr].Rmdir(input, output)
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
		err := self.servercoords[kmeta.Primary.CoordAddr].Rmdir(input, output)
		if err != nil || output.Status != fuse.OK {
			fmt.Println(input.Name, ": ", err, output.Status)
			panic(err)
		}
	}


	return nil
}

func (self *ServerCoordinator) Unlink(input *Unlink_input, output *Unlink_output) error {
	fmt.Println("Unlink: "+input.Name)
	kmeta, e := self.kc.GetData(input.Name)
	if e != nil { return e }

	if self.Addr == kmeta.Primary.CoordAddr {
		Lock(self, input.Name)

		e = self.kc.Remove(input.Name, kmeta)

		if e != nil {
			panic(e)
		}
		self.sfs.Unlink(input, output)
		self.kc.RemoveServerMeta(input.Name, self.Addr, false)
		for _, replica := range kmeta.Replicas {
			// get client
			client := self.serverfsm[replica.SFSAddr]
			e = client.Unlink(input, output)
			if e != nil {
				return e
			}
			self.kc.RemoveServerMeta(input.Name, replica.SFSAddr, true)
		}

		Unlock(self, input.Name)
	} else {
		client := self.servercoords[kmeta.Primary.CoordAddr]
		e = client.Unlink(input, output)
		if e != nil {
			panic(e)
		}
	}

	return nil
}

func (self *ServerCoordinator) Create(input *Create_input, output *Create_output) error {
	fmt.Println("Create:", input.Path)

	kmeta, e := self.kc.GetData(input.Path)
	if e == nil { // file already exists
		// pass
	} else if e.Error() == "Deleted boolean" {
		if self.Addr == kmeta.Primary.CoordAddr {
			Lock(self, input.Path)

			self.sfs.Create(input, output)
			_, e := self.kc.Create(input.Path, *output.Attr, kmeta.Deleted)
			if e != nil {
				return e
			}

			for _, replica := range kmeta.Replicas {
				// get client
				client := self.serverfsm[replica.SFSAddr]
				e = client.Create(input, output)
				if e != nil {
					return e
				}
			}

			Unlock(self, input.Path)
		} else {
			client := self.servercoords[kmeta.Primary.CoordAddr]
			e = client.Create(input, output)
			if e != nil {
				panic(e)
			}
		}
	} else {
		Lock(self, input.Path)
		self.sfs.Create(input, output)
		_, e := self.kc.Create(input.Path, *output.Attr, false)
		if e != nil {
			return e
		}
		Unlock(self, input.Path)
	}

	return nil
}

func (self *ServerCoordinator) FileRead(input *FileRead_input, output *FileRead_output) error {
	fmt.Println("Read -", "Path:", input.Path)
	self.sfs.FileRead(input, output)
	self.kc.Inc(input.Path, true)
	return nil
}

func (self *ServerCoordinator) FileWrite(input *FileWrite_input, output *FileWrite_output) error {

	fmt.Println("Write -", "Path:", input.Path, "Data:", input.Data)

	kmeta, e := self.kc.GetData(input.Path)
	if e != nil {
		return e
	}
	if self.Addr == kmeta.Primary.CoordAddr {
		Lock(self, input.Path)

		fmt.Println("Is the primary, path:",input.Path,"offset:",input.Off)
		self.sfs.FileWrite(input, output)

		for _, replica := range kmeta.Replicas {
			fmt.Println("Write on replica:", replica.SFSAddr)
			client := self.serverfsm[replica.SFSAddr]
			e = client.FileWrite(input, output)
			if e != nil {
				return fmt.Errorf("File Write Error %v\n", e)
			}
		}

		// increment on the correct addr
		kmeta.Attr = *output.Attr
		e = self.kc.IncAddr(input.Path, input.Faddr, kmeta, false)
		if e != nil {
			return e
		}

		Unlock(self, input.Path)
	} else {
		fmt.Println("Not primary, forwarding request to primary coordinator")
		client := self.servercoords[kmeta.Primary.CoordAddr]
		input.Kmeta = kmeta
		// tell to increment write on this server
		input.Faddr = self.SFSAddr
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

func Lock (self *ServerCoordinator, file string) {
	_ , ok := self.fileLocks[file]
	if !ok {
		self.fileLocks[file] = make(chan int, 1)
	}
	self.fileLocks[file]<-1
}

func Unlock (self *ServerCoordinator, file string) {
	<-self.fileLocks[file]
}

// assert that ServerCoordinator implements BackendFs
var _ BackendFs = new(ServerCoordinator)

/* ----------------------------------------------------------------------------
-----------------------Non BackendFs/ClientFs Functions------------------------
-----------------------------------------------------------------------------*/

func (self *ServerCoordinator) AddPathBackup(path, newCoordAddr string) error {
	input := &Open_input{Name: path}
	output := &Open_output{}
	e := self.servercoords[newCoordAddr].Open(input, output)
	if e != nil || output.Status != fuse.OK {
		fmt.Println("output status", output.Status)
		fmt.Println(e)
		panic(e)
	}

	return nil
}

// can be called from any coord?
func (self *ServerCoordinator) RemovePathBackup(path, backupSFSAddr string, currentReplicaDead bool) error {
	// get keeper data
	kmeta, e := self.kc.GetData(path)
	if e != nil {
		panic(e)
	}
	// check path is on backup
	_, exists := kmeta.Replicas[backupSFSAddr]
	if !exists {
		fmt.Println(backupSFSAddr)
		fmt.Println(kmeta.Replicas)
		panic("expected backup does not exists")
	}

	// remove backup
	if !currentReplicaDead {
		client := self.serverfsm[backupSFSAddr]
		input := &Unlink_input{Name: path}
		output := &Unlink_output{}
		e = client.Unlink(input, output)
		if e != nil || output.Status != fuse.OK {
			panic(e)
		}
	}
	self.kc.RemoveServerMeta(path, backupSFSAddr, true)

	// update keeper
	delete(kmeta.Replicas, backupSFSAddr)
	e = self.kc.Set(path, kmeta)
	if e != nil {
		panic(e)
	}

	// if no replicas we need to assign one
	if len(kmeta.Replicas) == 0 {
		fmt.Println(path, "has no backups!!!!!")
		// pick a new replica and add metadata to both the server meta and file meta
		replica := ""
		for addr, _ := range(self.servercoords) {
			// if the primary is not this address then assign it the replica
			if kmeta.Primary.CoordAddr != addr {
				replica = addr
			}
		}
		if replica != "" {
			// we have picked a replica so add metadata
			fmt.Println("picked replica", replica)
			self.AddPathBackup(path, replica)
		}
	}
	return nil

}

// should be called on new primary
func (self *ServerCoordinator) SwapPathPrimary(path string, currentPrimaryDead bool) error {
	// get keeper data
	kmeta, e := self.kc.GetData(path)
	if e != nil {
		panic(e)
	}

	// Chooose replacement
	var newPrimary ServerFileMeta
	maxWrite := -1
	for _, replica := range kmeta.Replicas {
		if currentPrimaryDead {
			if replica.WriteCount > maxWrite {
				maxWrite = replica.WriteCount
				newPrimary = replica
			}
		} else {
			if replica.WriteCount > maxWrite && replica.WriteCount > kmeta.Primary.WriteCount {
				maxWrite = replica.WriteCount
				newPrimary = replica
			}
		}
	}
	if (maxWrite < 0) {
		// failed to find replacement keep this one
		return nil
	}
	// perform swap
	if !currentPrimaryDead {
		oldPrimary := kmeta.Primary
		self.kc.RemoveServerMeta(path, oldPrimary.CoordAddr, false)
		self.kc.AddServerMeta(path, oldPrimary.SFSAddr, true)
		kmeta.Replicas[oldPrimary.SFSAddr] = oldPrimary
	}

	self.kc.RemoveServerMeta(path, newPrimary.SFSAddr, true)
	self.kc.AddServerMeta(path, newPrimary.CoordAddr, false)
	kmeta.Primary = newPrimary
	delete(kmeta.Replicas, newPrimary.SFSAddr)

	// update keeper
	e = self.kc.Set(path, kmeta)
	if e != nil {
		panic(e)
	}

	if len(kmeta.Replicas) == 0 {
		// need to assign a new replica for this
		fmt.Println(path, "has no backups!!!!!")
		// pick a new replica and add metadata to both the server meta and file meta
		replica := ""
		for addr, _ := range(self.servercoords) {
			// if the primary is not this address then assign it the replica
			if kmeta.Primary.CoordAddr != addr {
				replica = addr
			}
		}
		if replica != "" {
			// we have picked a replica so add metadata
			self.AddPathBackup(path, replica)
		}
	}

	return nil
}

func (self *ServerCoordinator) GetAddress(input *string, output *string) error {
	if *input == "server" {
		*output = self.SFSAddr
	} else if *input == "coord" {
		*output = self.Addr
	}
	return nil
}
