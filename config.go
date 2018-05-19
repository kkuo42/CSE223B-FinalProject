package geodfs

import (
)

type StoreConfig struct {
	Addr string
	Store Storage
	Ready chan<-bool
}
