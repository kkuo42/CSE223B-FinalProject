package proj

import (
	"log"
	"strings"
	"net/rpc"
	"encoding/gob"
)

type ClientFs struct {
	Addr string
	FullAddr string
	conn *rpc.Client
}

func NewClientFs(faddr string) *ClientFs {
	/* need to register nested structs of input/outputs */
	gob.Register(&CustomReadResultData{})
	addr := strings.Split(faddr, "_")[0]
	return &ClientFs{Addr: addr, FullAddr: faddr}
}

func (self *ClientFs) Connect() error {

	if (self.conn == nil) {
		conn, e := rpc.DialHTTP("tcp", self.Addr)
		if e != nil {
			log.Println("error connecting: ", e)
			return e
		}
		self.conn = conn
	}

	return nil
}

func (self *ClientFs) Open(input *Open_input, output *Open_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Open", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Open", input, output)
		}
	}

	return e
}

func (self *ClientFs) OpenDir(input *OpenDir_input, output *OpenDir_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.OpenDir", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.OpenDir", input, output)
		}
	}

	return e
}

func (self *ClientFs) GetAttr(input *GetAttr_input, output *GetAttr_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.GetAttr", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.GetAttr", input, output)
		}
	}

	return e
}

func (self *ClientFs) Unlink(input *Unlink_input, output *Unlink_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Unlink", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Unlink", input, output)
		}
	}

	return e
}

func (self *ClientFs) ReplicaUnlink(input *Unlink_input, output *Unlink_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.ReplicaUnlink", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.ReplicaUnlink", input, output)
		}
	}

	return e
}

func (self *ClientFs) Create(input *Create_input, output *Create_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Create", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Create", input, output)
		}
	}

	return e
}


func (self *ClientFs) FileRead(input *FileRead_input, output *FileRead_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.FileRead", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.FileRead", input, output)
		}
	}

	return e
}

func (self *ClientFs) FileWrite(input *FileWrite_input, output *FileWrite_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.FileWrite", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.FileWrite", input, output)
		}
	}

	return e
}

func (self *ClientFs) ReplicaFileWrite(input *FileWrite_input, output *FileWrite_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.ReplicaFileWrite", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.ReplicaFileWrite", input, output)
		}
	}

	return e
}

func (self *ClientFs) FileRelease(input *FileRelease_input, output *FileRelease_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.FileRelease", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.FileRelease", input, output)
		}
	}

	return e
}

func (self *ClientFs) Rename(input *Rename_input, output *Rename_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Rename", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Rename", input, output)
		}
	}

	return e
}

func (self *ClientFs) ReplicaRename(input *Rename_input, output *Rename_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.ReplicaRename", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.ReplicaRename", input, output)
		}
	}

	return e
}

func (self *ClientFs) Mkdir(input *Mkdir_input, output *Mkdir_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Mkdir", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Mkdir", input, output)
		}
	}

	return e
}

func (self *ClientFs) Rmdir(input *Rmdir_input, output *Rmdir_output) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.Rmdir", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.Rmdir", input, output)
		}
	}

	return e
}

func (self *ClientFs) GetAddress(input *string, output *string) error {
	e := self.Connect()
	if e != nil { return e }

	e = self.conn.Call("BackendFs.GetAddress", input, output)
	if e != nil {
		self.conn = nil
		e = self.Connect()
		if e == nil {
			e = self.conn.Call("BackendFs.GetAddress", input, output)
		}
	}

	return nil
}
// assert that ClientFs implements BackendFs
var _ BackendFs = new(ClientFs)


