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

type Entry struct {
	key   string
	value string
	ref   int64
}

type Store struct {
	listener net.Listener
	data     []Entry
	len      int
}

type StoreAPI int

var ssrv Store

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
	ssrv = Store{listener: listener}
	http.Serve(ssrv.listener, nil)

	if err != nil {
		log.Fatal("error serving: ", err)
	}
}
