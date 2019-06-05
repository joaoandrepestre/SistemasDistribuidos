package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type API int

type Message struct {
	MessageType int
	ProcessId   int
	ElectionId  int
}

func (api *API) receiveMsg(msg int, reply *int) error {
	/* var splitMsg []string

	splitMsg = strings.Split(msg, "|")
	msgType := splitMsg[0] */

	switch msg {
	case 1:
		fmt.Println("Eleição")
	case 2:
		fmt.Println("OK")
	case 3:
		fmt.Println("Líder")
	case 4:
		fmt.Println("Vivo")
	case 5:
		fmt.Println("Vivo_OK")
	}

	return nil
}

var id int
var electionId int
var leaderId int

func main() {
	api := new(API)
	rpc.Register(api)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	go http.Serve(l, nil)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply int
	//msg := Message{MessageType: 1, ProcessId: 3, ElectionId: 19}
	err = client.Call("API.receiveMsg", 1, &reply)
	if err != nil {
		log.Fatal("api error:", err)
	}
}

func userInterface() {
	for {
		fmt.Printf("Interface com o usuário\n")
		time.Sleep(time.Second * 1)
	}
}

func detectLeader() {
	for {
		fmt.Printf("Exporadicamente checa se o líder está ativo\n")
		time.Sleep(time.Second * 1)
	}
}
