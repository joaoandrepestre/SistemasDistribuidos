package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
)

/* type Message struct {
	MessageType int
	ProcessId   int
	ElectionId  int
}

type API int

func (api *API) ReceiveMsg(msg Message, reply *Message) error {
	var splitMsg []string

	splitMsg = strings.Split(msg, "|")
	msgType := splitMsg[0]

	switch msg.MessageType {
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
} */

var pid int
var eleicao_id int
var lider_id int

func main() {

	var numero_processos int
	var err error
	var port int
	var clients []*net.TCPConn

	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("Número de argumentos inválido. Forneça o número de processos e o ID do processo atual.")
		os.Exit(1)
	}
	numero_processos, err = strconv.Atoi(args[0])

	pid = os.Getpid()
	eleicao_id, err = strconv.Atoi(args[1])
	lider_id = numero_processos - 1

	var wg sync.WaitGroup
	wg.Add(1)

	port = 4000 + eleicao_id
	addr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go http.Serve(l, nil)

	for i := 0; i < numero_processos; i++ {
		if i != eleicao_id {
			addr, err = net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(4000+i))
			client, err := net.DialTCP("tcp", nil, addr)
			if err != nil {
				log.Fatal("dialing:", err)
			}
			clients = append(clients, client)
		}
	}

	/* client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	} */

	wg.Wait()
}
