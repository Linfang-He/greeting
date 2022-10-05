package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

type ServerConfigs struct {
	Servers []struct {
		ServerId int    `yaml:"serverId"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	} `yaml:"servers"`
}

func readServerConfigs(configPath string) ServerConfigs {
	f, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("could not read config file %s : %v", configPath, err)
	}
	scs := ServerConfigs{}
	err = yaml.Unmarshal(f, &scs)
	return scs
}

// As a listener
func listenForData(ch chan<- string, cType string, host string, port string) {
	fmt.Println("Starting " + cType + " server on connHost: " + host + ", connPort: " + port)
	l, err := net.Listen(cType, host + ":" + port)
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
}

func consolidateServerData(ch <-chan string, numOfServers int){
	numOfServersCompleted := 0
	for {
		if numOfServersCompleted == numOfServers - 1 {
			break
		}
		message := <-ch // receive data from channel
		fmt.Println(message)
		numOfServersCompleted += 1
	}
}

// As a sender
func sendAllData(scs ServerConfigs, myServerId int, allDone chan<- int) {
	done := make(chan int) // signal for one greetings sent
	defer close(done)

	count := 0 // count how many greetings has been sent
	for _, sc := range scs.Servers {
		if sc.ServerId != myServerId {
			message := "Greetings from server " + strconv.Itoa(myServerId)
			go sendData("tcp", sc.Host, sc.Port, message, done)
		}
	}
	for {
		if count == len(scs.Servers)-1 {
			allDone <- 1
			return
		}
		<-done // receive from each send success
		count = count + 1
	}
}

func sendData(cType string, host string, port string, message string, done chan<- int) {
	conn, err := net.Dial(cType, host + ":" + port)
	tm := 100 // wait at most 10s
	for err != nil {
		if tm == 0 {
			log.Panicln(err)
		}
		tm -= 1
		conn, err = net.Dial(cType, host + ":" + port)
		time.Sleep(100 * time.Millisecond) // 0.1s
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Panicln(err)
	}
	done <- 1 // signal of data already sent
}

func main() {
	// What is my serverId
	serverId, _ := strconv.Atoi(os.Args[1])

	// Read server configs from file
	scs := readServerConfigs(os.Args[2])

	// listen on a socket
	ch := make(chan string) // make sure receive data from all other servers
	defer close(ch)
	numOfServers := len(scs.Servers)
	go listenForData(ch, "tcp", scs.Servers[serverId].Host, scs.Servers[serverId].Port)
	
	// send greetings to others
	sendDone := make(chan int)  // make sure data is sent
	defer close(sendDone)
	go sendAllData(scs, serverId, sendDone)

	// consolidate data
	// will block until all data has been received
	consolidateServerData(ch, numOfServers)

	// close after data has been sent
	<-sendDone
}
