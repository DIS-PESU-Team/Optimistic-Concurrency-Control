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

// An entry is a state: (key, value, unique ref when written)
type Entry struct {
	key   string
	value string
	ref   int64
}

// A store is a list of entries
type Store struct {
	// Include a
	listener net.Listener
	data     []Entry
	len      int
}

// Global store containing list of entries
var ssrv Store

// Create a type to be able to register functions run by clients on the Store (RPC server)
type StoreAPI int

// Read the value of the key using the stores's API
// Return as a new entry object
func (a *StoreAPI) Read(key string, reply Entry) error {
	for _, entry := range ssrv.data {
		if entry.key == key {
			reply.key = entry.key
			reply.value = entry.value
			reply.ref = entry.ref
			break
		}
	}
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

	for i := 1; i <= n; i++ {
		temp = Entry{key: strconv.Itoa(i), value: "0", ref: time.Now().UnixNano()}
		ssrv.data = append(ssrv.data, temp)
	}

	api := new(StoreAPI)
	// Register the StoreAPI object which will be called from the Store (RPC server)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("error registering Handler API", err)
	}

	// Register an HTTP handler for RPC messages
	rpc.HandleHTTP()

	// Start listening to requests on port `os.Args[1]`
	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		log.Fatal("Listener error: ", err)
	}

	fmt.Printf("Serving RPC on port %s", os.Args[1])
	ssrv = Store{listener: listener}
	http.Serve(ssrv.listener, nil)

	if err != nil {
		log.Fatal("error serving: ", err)
	}
}
