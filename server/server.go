package main

import (
	"bytes"
	"fmt"
	"io"
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

// For RPC:
// - Functions need to be methods; must be uppercased to be exported
// - Must have two args, of exported or builtin type
// - The second arg must be a pointer - the result to client
// - Return type must be an error type
type Reply struct {
	Port string
}

// Registers the transaction handler creation method
type API int

// Create a new transaction handler for a client and return its port
func (a *API) GetHandlerPort(empty string, reply *Reply) error {
	pen := '\U0001F58A'

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
	fmt.Printf("%c  Registered a Client with Port: %s\n", pen, reply.Port)
	return nil
}

func main() {

	monitor := '\U0001F5A5'
	rocket := '\U0001F680'
	construction := '\U0001F6A7'
	//bulb := '\U0001F4A1'

	var reply Reply

	fmt.Printf("\n\t\t\t%c  Distributed Systems Project %c\n", monitor, monitor)
	fmt.Println("\tOptimistic Concurrency Control - Distributed Transaction Server")
	fmt.Println(" ------------------------------------------------------------------------------------")
	fmt.Printf("\t\t\t%c Starting up the Transaction Server\n\n", rocket)

	// Initiate the validator process at port 4041
	cmd := exec.Command("go", "run", "validator.go", "4041")

	// Attach the standard out to read what the command might print
	var stdBufferVal bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBufferVal)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err := cmd.Start()
	time.Sleep(1 * time.Second)
	fmt.Printf(" %c Started Validator on Port: 4041\n", construction)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Initiate the process with the key-value store at port 4042
	// Send the number of entries in the store as arg
	cmd = exec.Command("go", "run", "store.go", "4042", os.Args[1])

	// Attach the standard out to read what the command might print
	var stdBufferStr bytes.Buffer
	mw_str := io.MultiWriter(os.Stdout, &stdBufferStr)
	cmd.Stdout = mw_str
	cmd.Stderr = mw_str

	err = cmd.Start()
	time.Sleep(1 * time.Second)
	fmt.Printf(" %c Started Store on Port: 4042\n", construction)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Connect to the validator and store servers
	validator, err := rpc.DialHTTP("tcp", "localhost:4041")
	if err != nil {
		log.Fatal("Connection error to validator: ", err)
	}
	store, err := rpc.DialHTTP("tcp", "localhost:4042")
	if err != nil {
		log.Fatal("Connection error to store: ", err)
	}

	// Stop the validator and store processes in case of an exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		validator.Call("ValidatorAPI.Stop", "", reply)
		store.Call("StoreAPI.Stop", "", reply)
		os.Exit(1)
	}()

	// Initiate the server's API
	api := new(API)
	err = rpc.Register(api)
	if err != nil {
		log.Fatal("Error registering API: ", err)
	}

	rpc.HandleHTTP()

	// Initiate a listener
	listener, err := net.Listen("tcp", ":4040")
	if err != nil {
		log.Fatal("Listener error", err)
	}

	fmt.Printf(" %c Started Server on Port: 4040\n\n", construction)

	http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Error serving at 4040: ", err)
	}

}
