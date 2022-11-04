package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// The client makes asynchronous reads to the store. If the client dies, its transaction handler dies
// The transaction handler:
// - Records all read ops + timestamps
// - Records all write ops in a local cache

type HandlerAPI int
type Entry struct {
	Key   string
	Value string
	Ref   int64
}

type ReadSetEntry struct {
	ReadTrx  Entry
	ReadTime int64
}

type Txnstats struct {
	TotalTxns float32
	SuccTxns  float32
}

type WriteSetEntry struct {
	Key       string
	NewValue  string
	WriteTime int64
}

type Handler struct {
	Listener  net.Listener
	Store     *rpc.Client
	TotalTxns float32
	SuccTxns  float32

	// To record read ops, one entry: {Key, Value, PrevTimestamp}
	ReadSet []ReadSetEntry

	// To record write ops, one entry: {Key, Value, NewTimestamp}
	WriteSet []WriteSetEntry
}

// Global transaction handler for the client
var hsrv Handler

type Validator struct {
	Mu        sync.Mutex
	Validator *rpc.Client
}

var hvalidator Validator

// Data structure to send local read & write sets to the Validator
type RWSets struct {
	ClientRef string
	ReadSet   []ReadSetEntry
	WriteSet  []WriteSetEntry
	Commit    bool
}

// Method to read a value from the store using a key
func (a *HandlerAPI) Read(key string, reply *Entry) error {
	// Temp variable to store entry info corresponding to the key
	var storeReply Entry

	// If the key is in the local cache of data, return
	for _, entry := range hsrv.WriteSet {
		if entry.Key == key {
			reply.Key = entry.Key
			reply.Value = entry.NewValue
			reply.Ref = entry.WriteTime
			fmt.Printf("... Transaction (from cache): Read (Port: %s | Key: %s) - {%s, %s, %d} \n", os.Args[1], key, reply.Key, reply.Value, reply.Ref)
			return nil
		}
	}

	// If key not found in local cache, read from the global store
	hsrv.Store.Call("StoreAPI.Read", key, &storeReply)
	// Update the local cache with the entry for easier future reference
	hsrv.ReadSet = append(hsrv.ReadSet, ReadSetEntry{storeReply, time.Now().UnixMicro()})

	// Update the reply with the entry info obtained from the store
	reply.Key = storeReply.Key
	reply.Value = storeReply.Value
	reply.Ref = storeReply.Ref

	fmt.Printf("... Transaction: Read | Port: %s | {%s, %s, %d} \n", os.Args[1], reply.Key, reply.Value, reply.Ref)
	return nil
}

// Method to write a value to a given key
func (a *HandlerAPI) Write(newEntry Entry, reply *Entry) error {

	// Updates only the local cache with the new reference
	curr_time := time.Now().UnixMicro()
	hsrv.WriteSet = append(hsrv.WriteSet, WriteSetEntry{newEntry.Key, newEntry.Value, curr_time})

	fmt.Printf("... Transaction: Write | Port: %s | {%s, %s, %d} \n", os.Args[1], newEntry.Key, newEntry.Value, curr_time)
	return nil
}

// Commits all the write updates stored in the local write cache into the global store
func (a *HandlerAPI) Commit(empty string, reply *Entry) error {

	var valReply Entry

	// The validator process can only work on one commit at a time
	// Ensure consistency using a mutex
	hvalidator.Mu.Lock()
	defer hvalidator.Mu.Unlock()

	var rwset RWSets
	rwset.ReadSet = hsrv.ReadSet
	rwset.WriteSet = hsrv.WriteSet
	rwset.ClientRef = os.Args[1]
	rwset.Commit = true

	// Call the validator process
	hvalidator.Validator.Call("ValidatorAPI.Validate", &rwset, &valReply)

	hsrv.TotalTxns += 1

	if valReply.Key == "Success" {
		hsrv.SuccTxns += 1
	}

	fmt.Printf("... Transaction: Commit | Port: %s | Curr Stats: %f, %f \n", os.Args[1], hsrv.SuccTxns, hsrv.TotalTxns)

	//Empty the read and write sets
	hsrv.ReadSet = []ReadSetEntry{}
	hsrv.WriteSet = []WriteSetEntry{}

	return nil
}

func (a *HandlerAPI) ReadAll(empty *Entry, reply *Entry) error {
	hsrv.Store.Call("StoreAPI.ReadAll", empty, reply)
	return nil
}

func (a *HandlerAPI) GetStats(empty string, reply *Txnstats) error {
	reply.TotalTxns = hsrv.TotalTxns
	reply.SuccTxns = hsrv.SuccTxns
	fmt.Printf("... Transaction handler statistics (Port: %s) | %f, %f \n\n", os.Args[1], reply.TotalTxns, reply.SuccTxns)

	return nil
}

func (a *HandlerAPI) Stop(empty string, reply *Entry) error {
	hsrv.Listener.Close()
	time.Sleep(2 * time.Second)
	return nil
}

func main() {

	flag := '\U0001F3C1'

	// Get a connection to the store
	store, err := rpc.DialHTTP("tcp", "localhost:4042")
	if err != nil {
		log.Fatal("Connection to Store Error: ", err)
	}

	// Get a connection to the validator
	validator, err := rpc.DialHTTP("tcp", "localhost:4041")
	if err != nil {
		log.Fatal("Connection to Validator Error: ", err)
	}

	// Create an API for the transaction handler
	handler_api := new(HandlerAPI)
	err = rpc.Register(handler_api)
	if err != nil {
		log.Fatal("Error registering Handler API", err)
	}
	// Configure it to listen at the new port
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		log.Fatal("Error creating a listener for the handler: ", err)
	}

	fmt.Printf(" %c Connected to the transaction handler on port: %s\n", flag, os.Args[1])

	// Instantiate the transaction handler with the
	// connections to the listener and store
	hsrv = Handler{Listener: listener, Store: store, TotalTxns: 0, SuccTxns: 0}
	hvalidator = Validator{Validator: validator}
	http.Serve(hsrv.Listener, nil)
	if err != nil {
		log.Fatal("Error serving the transaction handler: ", err)
	}
}
