This code allows a file system on one machine be hosted on a port so that another process can use the file system with a FUSE mount.  

install go-fuse `go get github.com/hanwen/go-fuse/...`

## Usage
compile and install with `make` (NOTE: repository root folder must be named "proj")

To launch a server, MOUNTPOINT will be the filesystem sent,
```
fs-server MOUNTPOINT
```

To launch a client, MOUNTPOINT will be where the filesystem received,
```
fs-front MOUNTPOINT
```
To unmount a client, stop the `fs-front` and `fusermount -u MOUNTPOINT`.

MOUNTPOINT should be a folder, but can be named whatever.

### example:
Two directories exist, "from" and "to". There are files in the "from" directory and none in "to". Running `fs-server from` then `fs-front to` will allow the forwards the files in "from" to also appear in "to". 

## Details 
High level description of client/server layers

### frontend
The frontend is a `pathfs.FileSystem` that is used as the FUSE interface.  Its file system is intialized with a `pathfs.defaultFileSystem` so that it will return errors for all the non override functions.  The frontend overides the `FileSystem` interface functions so that they use the `BackendFs`. 

### client
`ClientFs` implements `BackendFs` such that all calls are forwarded to a server using RPC

### backend
`BackendFs` is the RPC interface.

`ServerFs` implements `BackendFs` that uses the loopback filesystem to make system calls on the servers disk

### loopback
`CustomLoopbackFileSystem` is a FUSE filesystem that shunts all request to an underlying file system. (It does not need to be a fuse file system but this had all the implemented syscalls it made sense to use.)

### misc
`CustomLoopbackFile` and `CustomReadResultData` are %99 the same as their respective structs in the go-fuse repo but were needed to be modified slightly so that file wrapper that make syscalls are not sent to a frontend.  `FrontendFile` is a file wrapper created on the frontend which forwards requests through the BackendFs to a `CustomLoopbackFile` saved in the `ServerFs` so that syscall operations are performed on the server 
