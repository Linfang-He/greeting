package main

import (
	"fmt"
	"net"
	"os"
)

// As a listener
func listenForData(ch chan<- string, cType string, host string, port string) {
	fmt.Println("Starting " + cType + " server on connHost: " + host + ", connPort: " + port)
	l, err := net.Listen(cType, host+":"+port)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client " + conn.RemoteAddr().String() + " connected.")
		go handleConnection(conn, ch)
	}
}

func handleConnection(conn net.Conn, ch chan<- string) {
	buf := make([]byte, 100)
	bytes, err := conn.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conn::Read: err %v\n", err)
		os.Exit(1)
	}
	greeting := string(buf[0:bytes])
	ch <- greeting
	conn.Close()
}

func consolidateServerData(ch <-chan string, numOfClients int) {
	numOfClientsCompleted := 0
	for {
		if numOfClientsCompleted == numOfClients {
			break
		}
		message := <-ch // receive data from channel
		fmt.Println(message)
		numOfClientsCompleted++
	}
}

func main() {
	// Read server configs
	host := os.Args[1]
	port := os.Args[2]

	// listen on a socket
	ch := make(chan string)
	defer close(ch)
	go listenForData(ch, "tcp", host, port)

	// consolidate data
	// will block until all data has been received
	numOfClients := 7
	consolidateServerData(ch, numOfClients)
}