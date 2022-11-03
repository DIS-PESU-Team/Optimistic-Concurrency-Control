package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Validator struct {
	listener       net.Listener
	store          *rpc.Client
	numTxns        int
	successfulTxns int
}

type Entry struct {
	Key   string
	Value string
	Ref   int64
}

type ReadSetEntry struct {
	ReadTrx  Entry
	ReadTime int64
}

type WriteSetEntry struct {
	Key       string
	NewValue  string
	WriteTime int64
}

type RWSets struct {
	ClientRef string
	ReadSet   []ReadSetEntry
	WriteSet  []WriteSetEntry
	Commit    bool
}
type Valid struct {
	ReadValid bool
}

type ValidatorAPI int

var vsrv Validator

func (a *ValidatorAPI) Validate(rwset *RWSets, reply *Entry) error {
	var storeReply Entry

	fmt.Println("\n\tTransaction for Client at Port: ", rwset.ClientRef)
	rwset.Commit = true
	// Check if all read ops are still valid
	for _, read_entry := range rwset.ReadSet {
		vsrv.store.Call("StoreAPI.Read", read_entry.ReadTrx.Key, &storeReply)
		if storeReply.Ref != read_entry.ReadTrx.Ref {
			fmt.Println("ABORT. Transaction commit failed due to differing reads.")
			fmt.Println(storeReply.Ref, read_entry.ReadTrx.Ref)
			rwset.Commit = false
			break
		}
	}

	// If all read ops are valid, update all pending writes
	if rwset.Commit == true {
		fmt.Println("\t...All read operations are valid!")
		for _, write_entry := range rwset.WriteSet {
			vsrv.store.Call("StoreAPI.Write", &write_entry, &storeReply)
			fmt.Println("\t...Updated entry with key: ", write_entry.Key)
		}
	}

	return nil
}

func (a *ValidatorAPI) Stop(empty string, reply *Entry) error {
	vsrv.listener.Close()
	time.Sleep(time.Second * 1)
	return nil
}

func main() {

	// Register the Validator's methods
	api := new(ValidatorAPI)
	err := rpc.Register(api)
	if err != nil {
		log.Fatal("Error registering Handler API", err)
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+os.Args[1])

	if err != nil {
		log.Fatal("Listener error at Validator:", err)
	}

	// Get a connection to the store
	store, str_err := rpc.DialHTTP("tcp", "localhost:4042")
	if str_err != nil {
		log.Fatal("Connection to Store Error: ", str_err)
	}

	// fmt.Printf("serving rpc on port %s", os.Args[1])
	vsrv = Validator{listener: listener, store: store}
	http.Serve(vsrv.listener, nil)

	if err != nil {
		log.Fatal("Error serving at Validator:", err)
	}
}
