package main

import (
    "strings"
    "fmt"
    "os"
    "flag"
    "bufio"
    "geodfs/store"
)

func runCmd(sc *store.StorageClient, args []string) bool {
    cmd := args[0]
    switch cmd {
    case "stat":
        res, e := sc.Stat(args[1])
        if e != nil {
            fmt.Println("ERRORORORORORO")
            return true
        } else {
            fmt.Println(res)
        }
        return false
    case "read":
        res, e := sc.Read(args[1])
        if e != nil {
            fmt.Println("ERRORORORORORO")
            return true
        } else {
            fmt.Println(string(res.Data))
        }
        return false
    case "exit":
        return true
    default:
        fmt.Println("default")
        return false
    }
}

func fields(s string) []string {
    return strings.Fields(s)
}

func runPrompt(sc *store.StorageClient) {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("> ")

    for scanner.Scan() {
        line := scanner.Text()
        args := fields(line)
        if len(args) > 0 {
            if runCmd(sc, args) {
                break
            }
            fmt.Print("> ")
        }
    }

    e := scanner.Err()
    if e != nil {
        panic(e)
    }
}

func main() {
    flag.Parse()
    args := flag.Args()
    if len(args) < 2 {
        fmt.Println("usage: geodfs <client/server> <addr>")
        os.Exit(1)
    }

    addr := args[1]
    if args[0] == "server" {
        e := store.ServeStorage(addr, "/Users/cjpais/go/src/geodfs/storagedir/")
        if e != nil {
            fmt.Println("somethings up yo")
        }
    } else if args[0] == "client" {
        sc := store.NewStorageClient(addr)
        e := sc.Init()
        if e != nil {
            fmt.Println("error initializing client", e)
        }
        runPrompt(sc)
    } else {
        fmt.Println("usage: geodfs <client/server> <addr>")
        os.Exit(1)
    }
}
