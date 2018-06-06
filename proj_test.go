package proj_test

import (
	"testing"
	
	"fmt"
	"time"
	"strconv"
	"os"
	"os/exec"
	"log"
	"bytes"
	"proj"
	"github.com/hanwen/go-fuse/fuse"
)

func Start(i int) (retval []*proj.ServerCoordinator) {
	setupZK()
	
	os.RemoveAll("data")
	os.Mkdir("data", os.ModePerm)
	for j := 0; j<i; j++ {
		retval = append(retval, setupPair(j))
	}

	return retval
}

func Stop(i int) {
	fmt.Println()
	cmd := exec.Command("zookeeper-3.4.12/bin/zkServer.sh", "stop")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(out.String())

	for j := 0; j<i; j++ {
		stopPair(j)
	}
}

func setupZK() {
	os.RemoveAll("zkdata")
	os.Remove("zookeeper.out")
	os.Mkdir("zkdata", os.ModePerm)

	cmd := exec.Command("zookeeper-3.4.12/bin/zkServer.sh", "start")
	var out bytes.Buffer
	cmd.Stdout = &out	
	fmt.Println()
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(out.String())
}


func setupPair(i int) *proj.ServerCoordinator {

	insert := strconv.Itoa(i)

	sharepoint := "data/from" + insert
	os.Mkdir(sharepoint, os.ModePerm)
	coordaddr := "localhost:950" + insert
	fsaddr := "localhost:960" + insert
	fmt.Println()
	retvalChannel := make(chan *proj.ServerCoordinator)

    fsserver := proj.NewFsserver(sharepoint, coordaddr, fsaddr, retvalChannel)
    go fsserver.Run()
    retval := <- retvalChannel
    time.Sleep(time.Second)

	// backaddr := coordaddr
	// mountpoint := "data/to" + insert
	// os.Mkdir(mountpoint, os.ModePerm)
	// fmt.Println()
 //    go fsfront.Run(backaddr, mountpoint)
 //    time.Sleep(time.Second)
    return retval
}

func stopPair(i int) {
	cmd := exec.Command("fusermount", "-u", "data/to"+strconv.Itoa(i))
	var out bytes.Buffer
	cmd.Stdout = &out	
	cmd.Run()
}

func createFile(coord *proj.ServerCoordinator, path string) {
	input := &proj.Create_input{Path: path}
	output := &proj.Create_output{}
	e := coord.Create(input, output)
	if e != nil || output.Status != fuse.OK {
		panic(e)
	}
}


func TestAddPathBackup(t *testing.T) {
	pairs := 3
	coordinators := Start(pairs)
	path := "file0"
	createFile(coordinators[0], path)
	coordinators[0].AddPathBackup(path, coordinators[2].Addr)
	// coordinators[0].RemovePathBackup(path, coordinators[2].SFSAddr)

    Stop(pairs)
}
