package main

import (
    "proj"
    "fmt"
)

func main() {	
	tester := proj.NewTester(3)
	defer tester.Stop()

	filename := "file0"

	tester.CreateFile(0, filename)
	tester.AssertFileExists(tester.Coord[0].Path + "/" + filename)
	
	tester.Coord[0].AddPathBackup(filename, tester.Coord[2].Addr)
	tester.AssertFileExists(tester.Coord[2].Path + "/" + filename)

	tester.Coord[0].RemovePathBackup(filename, tester.Coord[2].SFSAddr)
	tester.AssertFileNotExists(tester.Coord[2].Path + "/" + filename)


	tester.Coord[0].AddPathBackup(filename, tester.Coord[2].Addr)
	tester.Coord[2].SwapPathPrimary(filename, false)
	tester.Coord[2].RemovePathBackup(filename, tester.Coord[0].SFSAddr)	
	tester.AssertFileExists(tester.Coord[2].Path + "/" + filename)
	tester.AssertFileNotExists(tester.Coord[0].Path + "/" + filename)

	fmt.Println("\nPASSED: data_mvmt")
}