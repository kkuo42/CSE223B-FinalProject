This code allows a file system on one machine be hosted on a port so that another process can use the file system with a FUSE mount. 

install go-fuse `go get github.com/hanwen/go-fuse/...`

## Usage
compile and install with `make` (NOTE: repository root folder must be named "proj")

To launch a server, MOUNTPOINT will be the filesystem sent,
```
fs-server SHAREPOINT
```

To launch a client, MOUNTPOINT will be where the filesystem received,
```
fs-front MOUNTPOINT
```
To unmount a client, stop the `fs-front` and `fusermount -u MOUNTPOINT`.

`SHAREPOINT` and `MOUNTPOINT` are folders and can be named whatever.

### example:
Two directories exist, `from` and `to`. There are files in the "from" directory and there are none in "to". Running `fs-server from` then `fs-front to` will forwards the files in `from` into `to`. 

## Details 
High level description of client/server layers

### frontend
The `Frontend` is a `pathfs.FileSystem` and `FrontendFile` is a `nodefs.File`. These are used as to implementet the abstracted `go-fuse` FUSE interface. The `Frontend` file system is intialized with a `pathfs.defaultFileSystem` so that it will return errors for all the non override functions. The frontend overides the `FileSystem` interface functions so that they use the `BackendFs`. 

### client
`ClientFs` implements `BackendFs` such that all calls are forwarded to a server using RPC

### backend
`BackendFs` is the RPC interface.

`ServerFs` implements `BackendFs` and uses the `CustomLoopbackFileSystem` to make system calls on the servers disk

### loopback
`CustomLoopbackFileSystem` is a FUSE filesystem that shunts all request to an underlying file system. (It does not need to be a fuse file system but this had all the implemented syscalls it made sense to use.)

### misc
`CustomLoopbackFile` and `CustomReadResultData` are %99 the same as their respective structs in the go-fuse repo but were needed to be modified slightly so that file wrapper that make syscalls are not sent to a frontend. `FrontendFile` is a file wrapper created on the frontend which forwards requests through the `BackendFs` to a `CustomLoopbackFile` saved in the `ServerFs` so that syscall operations are performed on the server 
