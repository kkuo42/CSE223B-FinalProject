package proj

import (
	"net/rpc"
    "encoding/gob"
)

type ClientFs struct {
	addr string
	conn *rpc.Client
}

func NewClientFs(addr string) ClientFs {
	gob.Register(&CustomReadResultData{})
	return ClientFs{addr: addr}
}

func (self *ClientFs) Connect() error {

	if (self.conn == nil) {
		conn, e := rpc.DialHTTP("tcp", self.addr)
		if e != nil {
			return e
		}
		self.conn = conn
	}

	return nil
}

func (self *ClientFs) Open(input *Open_input, output *Open_output) error {
	e := self.Connect()
	if e != nil {
		return e
	}

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
	if e != nil {
		return e
	}

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
	if e != nil {
		return e
	}

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

func (self *ClientFs) FileRead(input *FileRead_input, output *FileRead_output) error {
	e := self.Connect()
	if e != nil {
		return e
	}

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
// assert that ClientFs implements BackendFs
var _ BackendFs = new(ClientFs)