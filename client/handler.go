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
	readTrx  Entry
	readTime int64
}

type WriteSetEntry struct {
	Key       string
	newValue  string
	writeTime int64
}

type Handler struct {
	listener net.Listener
	store    *rpc.Client

	// To record read ops, one entry: {Key, Value, PrevTimestamp}
	readSet []ReadSetEntry

	// To record write ops, one entry: {Key, Value, NewTimestamp}
	writeSet []WriteSetEntry
}

// Global transaction handler for the client
var hsrv Handler

type Validator struct {
	mu        sync.Mutex
	validator *rpc.Client
}

var hvalidator Validator

// Data structure to send local read & write sets to the Validator
type RWSets struct {
	ReadSet  []ReadSetEntry
	WriteSet []WriteSetEntry
	Commit   bool
}

// Method to read a value from the store using a key
func (a *HandlerAPI) Read(key string, reply *Entry) error {
	// Temp variable to store entry info corresponding to the key
	var storeReply Entry

	// If the key is in the local cache of data, return
	for _, entry := range hsrv.writeSet {
		if entry.Key == key {
			reply.Key = entry.Key
			reply.Value = entry.newValue
			reply.Ref = entry.writeTime
			fmt.Printf("... Transaction (from cache): Read (Port: %s | Key: %s) - {%s, %s, %d} \n", os.Args[1], key, reply.Key, reply.Value, reply.Ref)
			return nil
		}
	}

	// If key not found in local cache, read from the global store
	hsrv.store.Call("StoreAPI.Read", key, &storeReply)
	// Update the local cache with the entry for easier future reference
	hsrv.readSet = append(hsrv.readSet, ReadSetEntry{storeReply, time.Now().UnixMicro()})

	// Update the reply with the entry info obtained from the store
	reply.Key = storeReply.Key
	reply.Value = storeReply.Value
	reply.Ref = storeReply.Ref

	fmt.Printf("... Transaction: Read (Port: %s | Key: %s) - {%s, %s, %d} \n", os.Args[1], key, reply.Key, reply.Value, reply.Ref)
	return nil
}

// Method to write a value to a given key
func (a *HandlerAPI) Write(newEntry Entry, reply *Entry) error {

	// Updates only the local cache with the new reference
	hsrv.writeSet = append(hsrv.writeSet, WriteSetEntry{newEntry.Key, newEntry.Value, time.Now().UnixMicro()})

	fmt.Printf("... Transaction: Write (Port: %s | Key: %s)\n", os.Args[1], newEntry.Key)
	return nil
}

// Commits all the write updates stored in the local write cache into the global store
func (a *HandlerAPI) Commit(empty string, reply *Entry) error {
	var rwset RWSets
	hvalidator.mu.Lock()
	defer hvalidator.mu.Unlock()

	rwset.ReadSet = hsrv.readSet
	rwset.WriteSet = hsrv.writeSet

	// Call the validator process
	hvalidator.validator.Call("ValidatorAPI.Validate", " ", &rwset)
	fmt.Printf("... Transaction: Commit (Port: %s) \n", os.Args[1])

	return nil
}

func (a *HandlerAPI) Stop(empty string, reply *Entry) error {
	hsrv.listener.Close()
	time.Sleep(1 * time.Second)
	fmt.Printf("... Transaction handler stopped (Port: %s)\n\n", os.Args[1])
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
	hsrv = Handler{listener: listener, store: store}
	hvalidator = Validator{validator: validator}
	http.Serve(hsrv.listener, nil)
	if err != nil {
		log.Fatal("Error serving the transaction handler: ", err)
	}
}
