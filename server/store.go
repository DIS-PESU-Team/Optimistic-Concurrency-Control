package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

// Clients do not access the store directly. A transaction handler is created
// per client that holds the store and process id of the validator process.

// An entry is a state: (Key, Value, unique Ref when written)
type Entry struct {
	Key   string
	Value string
	Ref   int64
}

// A store is a list of entries
type Store struct {
	listener net.Listener
	data     []Entry
	len      int
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

// Global store containing list of entries
var ssrv Store

// Create a type to be able to register functions run by clients on the Store (RPC server)
type StoreAPI int

// Read the Value of the Key using the stores's API
// Return as a new entry object
func (a *StoreAPI) Read(Key string, reply *Entry) error {
	for _, entry := range ssrv.data {
		if entry.Key == Key {
			reply.Key = entry.Key
			reply.Value = entry.Value
			reply.Ref = entry.Ref
			break
		}
	}
	return nil
}

func (a *StoreAPI) Write(write_entry *WriteSetEntry, reply *Entry) error {
	for i, entry := range ssrv.data {
		if entry.Key == write_entry.Key {
			ssrv.data[i].Value = write_entry.NewValue
			ssrv.data[i].Ref = write_entry.WriteTime
		}
	}
	return nil
}

func (a *StoreAPI) ReadAll(empty *Entry, reply *Entry) error {
	fmt.Println()
	for _, entry := range ssrv.data {
		fmt.Println(entry.Key, entry.Value, entry.Ref)
	}
	fmt.Println()
	return nil
}

func (a *StoreAPI) Stop(empty string, reply *Entry) error {
	ssrv.listener.Close()
	time.Sleep(time.Second * 1)
	return nil
}

func main() {

	n, err := strconv.Atoi(os.Args[2])
	var temp Entry

	api := new(StoreAPI)
	// Register the StoreAPI object which will be called from the Store (RPC server)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("Error registering Handler API", err)
	}

	// Register an HTTP handler for RPC messages
	rpc.HandleHTTP()

	// Start listening to requests on port `os.Args[1]`
	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	// fmt.Printf("Serving RPC on port %s", os.Args[1])
	ssrv = Store{listener: listener}

	for i := 1; i <= n; i++ {
		temp = Entry{Key: strconv.Itoa(i), Value: "0", Ref: time.Now().UnixNano()}
		ssrv.data = append(ssrv.data, temp)
	}

	http.Serve(ssrv.listener, nil)
	if err != nil {
		log.Fatal("Error serving at Store: ", err)
	}
}
