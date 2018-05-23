package proj

import (
	"github.com/hanwen/go-fuse/fuse"
)

// this struct is to maintain all information relative to the keeper
type Keeper struct {
	Attr fuse.Attr
	Primary string
	Replicas []string
}
