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
	numTxns        int
	successfulTxns int
}

type Entry struct {
	key   string
	value string
	ref   int64
}

type RWSets struct {
	readSet  []Entry
	writeSet []Entry
	commit   bool
}

type ValidatorAPI int

var vsrv Validator

func (a *ValidatorAPI) Validate(empty string, reply *Entry) error {

	return nil
}

func (a *ValidatorAPI) Stop(empty string, reply *Entry) error {
	vsrv.listener.Close()
	time.Sleep(time.Second * 1)
	return nil
}

func main() {
	api := new(ValidatorAPI)
	err := rpc.Register(api)
	if err != nil {
		log.Fatal("error registering Handler API", err)
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+os.Args[1])

	if err != nil {
		log.Fatal("Listener error", err)
	}
	fmt.Printf("serving rpc on port %s", os.Args[1])
	vsrv = Validator{listener: listener}
	http.Serve(vsrv.listener, nil)

	if err != nil {
		log.Fatal("error serving: ", err)
	}
}
