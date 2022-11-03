package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Reply struct {
	Port  string
	value string
}

type Entry struct {
	Key   string
	Value string
	Ref   int64
}

func simulate(handler *rpc.Client, reads int, writes int, startIndex int, rangeIndex int) {
	var ind int
	var reply Entry

	for reads > 0 && writes > 0 {
		rw := rand.Float64()
		if reads == 0 {
			rw = 0
		} else if writes == 0 {
			rw = 1
		}

		if rw > 0.5 {
			fmt.Println("Read")
			ind = startIndex + rand.Intn(rangeIndex)
			handler.Call("HandlerAPI.Read", strconv.Itoa(ind), &reply)
			reads -= 1
		} else {
			fmt.Println("Write")
			ind = startIndex + rand.Intn(rangeIndex)
			reply.Value = strconv.Itoa(rand.Intn(100))
			handler.Call("HandlerAPI.Write", strconv.Itoa(ind), &reply)
			writes -= 1
		}

		time.Sleep(1000 * time.Microsecond)
	}

	fmt.Println("Txn Done")
	handler.Call("HandlerAPI.Commit", "", reply)
	time.Sleep(1 * time.Second)
}

func main() {
	var reply Reply
	laptop := '\U0001F4BB'
	pushpin := '\U0001F4CC'
	hammer := '\U0001F528'
	check := '\U00002705'
	//email := '\U00002709'

	fmt.Printf("\n%c Started a New Client\n", laptop)
	fmt.Printf("\n%c Getting a new port for the handler\n", pushpin)
	server, err := rpc.DialHTTP("tcp", "localhost:4040")
	server.Call("API.GetHandlerPort", " ", &reply)
	time.Sleep(2 * time.Second)
	fmt.Printf("%c Starting a new transaction handler at Port %s\n", pushpin, reply.Port)

	// Start up the new transaction handler process
	cmd := exec.Command("go", "run", "handler.go", reply.Port)
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Connect to the newly created transaction handler process
	fmt.Printf("%c Connecting to the transaction handler...\n", pushpin)
	time.Sleep(2 * time.Second)
	handler, err := rpc.DialHTTP("tcp", "localhost:"+reply.Port)
	if err != nil {
		log.Fatal("Connection error to transaction handler: ", err)
	}

	reads, err := strconv.Atoi(os.Args[1])
	writes, err := strconv.Atoi(os.Args[2])
	startIndex, err := strconv.Atoi(os.Args[3])
	rangeIndex, err := strconv.Atoi(os.Args[4])
	totalTime, err := strconv.Atoi(os.Args[5])

	beg := time.Now().UnixMilli()
	var cur int64
	fmt.Printf("%c Beginning Transactions...\n\n", hammer)
	for {
		simulate(handler, reads, writes, startIndex, rangeIndex)
		cur = time.Now().UnixMilli()
		if (cur - beg) > int64(totalTime*int(time.Microsecond)) {
			break
		}
	}

	handler.Call("HandlerAPI.Stop", " ", nil)
	fmt.Printf("\n%c All transactions complete. Shutting down the Client.\n", check)

}
