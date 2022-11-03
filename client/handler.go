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

type Entry struct {
	Key   string
	Value string
	Ref   int64
}

type Handler struct {
	listener net.Listener
	store    *rpc.Client
	data     []Entry
	readSet  []Entry
	writeSet []Entry
}

type Validator struct {
	mu        sync.Mutex
	validator *rpc.Client
}

type RWSets struct {
	ReadSet  []Entry
	WriteSet []Entry
	Commit   bool
}

type HandlerAPI int

var hsrv Handler
var hvalidator Validator

func (a *HandlerAPI) Read(key string, reply *Entry) error {
	var storeReply Entry
	for _, entry := range hsrv.data {
		if entry.Key == key {
			reply.Key = entry.Key
			reply.Value = entry.Value
			return nil
		}
	}
	hsrv.store.Call("StoreAPI.Read", key, &storeReply)
	hsrv.readSet = append(hsrv.readSet, Entry{storeReply.Key, storeReply.Value, storeReply.Ref})
	reply.Key = storeReply.Key
	reply.Value = storeReply.Value
	fmt.Println("Read")
	return nil
}

func (a *HandlerAPI) Write(key string, reply *Entry) error {
	hsrv.writeSet = append(hsrv.writeSet, Entry{key, reply.Value, time.Now().UnixMicro()})
	fmt.Println("Write")
	return nil
}

func (a *HandlerAPI) Commit(empty string, reply *Entry) error {
	var rwset RWSets
	hvalidator.mu.Lock()
	defer hvalidator.mu.Unlock()

	rwset.ReadSet = hsrv.readSet
	rwset.WriteSet = hsrv.writeSet

	hvalidator.validator.Call("ValidatorAPI.Validate", " ", &rwset)
	fmt.Println("Commit")

	return nil
}

func (a *HandlerAPI) Stop(empty string, reply *Entry) error {
	hsrv.listener.Close()
	time.Sleep(1 * time.Second)
	fmt.Println("STOP")
	return nil
}

func main() {

	store, err := rpc.DialHTTP("tcp", "localhost:4042")

	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	validator, err := rpc.DialHTTP("tcp", "localhost:4041")

	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	api := new(HandlerAPI)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("error registering Handler API", err)
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+os.Args[1])

	if err != nil {
		log.Fatal("Listener error", err)
	}
	fmt.Printf("serving rpc on port %s", os.Args[1])
	hsrv = Handler{listener: listener, store: store}
	hvalidator = Validator{validator: validator}
	http.Serve(hsrv.listener, nil)

	if err != nil {
		log.Fatal("error serving: ", err)
	}
}
