package main

import (
	"bytes"
	"fmt"
	"io"
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

// Simulate one transaction (atomic action - set of reads and writes)
func simulate(handler *rpc.Client, reads int, writes int, startIndex int, rangeIndex int) {
	var ind int
	var writeEntry Entry
	var reply Entry

	for reads > 0 || writes > 0 {

		// Randomly choose whether to read or write
		rw := rand.Float64()
		if reads == 0 {
			rw = 0
		} else if writes == 0 {
			rw = 1
		}

		if rw > 0.5 {
			ind = startIndex + rand.Intn(rangeIndex)
			handler.Call("HandlerAPI.Read", strconv.Itoa(ind), &reply)
			reads -= 1
		} else {
			writeEntry.Key = strconv.Itoa(startIndex + rand.Intn(rangeIndex))
			writeEntry.Value = strconv.Itoa(20 + rand.Intn(100))
			writeEntry.Ref = time.Now().UnixMicro()
			handler.Call("HandlerAPI.Write", writeEntry, &reply)
			writes -= 1
		}
		time.Sleep(1 * time.Second)
	}

	// Commit the transaction's result to the global store
	handler.Call("HandlerAPI.Commit", "", &reply)
	handler.Call("HandlerAPI.ReadAll", &reply, &reply)
	fmt.Printf("... Complete\n\n")
	time.Sleep(1 * time.Second)
}

func main() {
	var reply Reply
	laptop := '\U0001F4BB'
	pushpin := '\U0001F4CC'
	hammer := '\U0001F528'
	check := '\U00002705'
	placard := '\U0001FAA7'
	//email := '\U00002709'

	fmt.Printf("\n\t\t%c New Client Instantiated %c \n", laptop, laptop)

	reads, err := strconv.Atoi(os.Args[1])
	writes, err := strconv.Atoi(os.Args[2])
	startIndex, err := strconv.Atoi(os.Args[3])
	rangeIndex, err := strconv.Atoi(os.Args[4])
	clientRunTime, err := strconv.Atoi(os.Args[5])

	fmt.Printf("\n\t%c Runtime: %d | # Reads/Txn: %d | # Writes/Txn: %d \n", placard, clientRunTime*int(time.Microsecond), reads, writes)

	fmt.Printf("\n %c Getting a new port for a new transaction handler...\n", pushpin)
	server, err := rpc.DialHTTP("tcp", "localhost:4040")
	server.Call("API.GetHandlerPort", " ", &reply)
	time.Sleep(2 * time.Second)
	fmt.Printf(" %c Starting the transaction handler at port %s...\n", pushpin, reply.Port)

	// Start up the new transaction handler process
	cmd := exec.Command("go", "run", "handler.go", reply.Port)

	// Attach the standard out to read what the command might print
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting transaction handler", err.Error())
		return
	}

	// Connect to the newly created transaction handler process
	time.Sleep(2 * time.Second)
	handler, err := rpc.DialHTTP("tcp", "localhost:"+reply.Port)
	if err != nil {
		fmt.Println("Connection error to transaction handler: ", err)
	}

	fmt.Printf(" %c Beginning transactions...\n\n", hammer)

	// Simulate all transactions for the client
	var curr_time int64
	start_time := time.Now().UnixMilli()
	for {
		// Simulate one transaction
		simulate(handler, reads, writes, startIndex, rangeIndex)

		// Ensure the time limit for client running transactions
		curr_time = time.Now().UnixMilli()
		if (curr_time - start_time) > int64(clientRunTime*int(time.Microsecond)) {
			break
		}
	}

	handler.Call("HandlerAPI.Stop", " ", nil)
	fmt.Printf("\n %c All transactions complete.\n %c Shutting down the client on port %s.\n", check, check, reply.Port)

}
