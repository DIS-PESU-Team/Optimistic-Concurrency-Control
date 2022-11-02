package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Reply struct {
	Port string
}

type API int

func (a *API) GetHandlerPort(empty string, reply *Reply) error {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	reply.Port = strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	fmt.Println("Registered a Client with Port: ", reply.Port)
	return nil
}

func main() {

	var reply Reply

	cmd := exec.Command("go", "run", "validator.go", "4041")
	err := cmd.Start()
	time.Sleep(1 * time.Second)
	fmt.Println("Started Validator on Port: 4041")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmd = exec.Command("go", "run", "store.go", "4042", os.Args[1])
	err = cmd.Start()
	time.Sleep(1 * time.Second)
	fmt.Println("Started Store on Port: 4042")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	validator, err := rpc.DialHTTP("tcp", "localhost:4041")
	store, err := rpc.DialHTTP("tcp", "localhost:4042")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		validator.Call("ValidatorAPI.Stop", "", reply)
		store.Call("StoreAPI.Stop", "", reply)
		os.Exit(1)
	}()

	api := new(API)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("error registering API", err)
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":4040")

	if err != nil {
		log.Fatal("Listener error", err)
	}
	fmt.Println("Started Server on Port: 4040")
	http.Serve(listener, nil)

	if err != nil {
		log.Fatal("error serving: ", err)
	}

}
