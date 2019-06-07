package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var pid int
var eleicao_id int
var lider_id int
var numero_processos int
var listener *net.TCPListener
var clients []*net.TCPConn

func main() {

	var wg sync.WaitGroup
	wg.Add(1)
	var port int

	// Recupera os argumentos da linha de comando
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("Número de argumentos inválido. Forneça o número de processos e o ID do processo atual.")
		os.Exit(1)
	}
	numero_processos, err := strconv.Atoi(args[0])
	eleicao_id, err = strconv.Atoi(args[1])
	if eleicao_id < 0 || eleicao_id >= numero_processos {
		log.Fatal("Id para eleição inválido. Forneça um ID entre 0 e ", numero_processos-1)
		os.Exit(1)
	}

	// Inicia outras variáveis globais
	pid = os.Getpid()
	lider_id = numero_processos - 1 // Líder inicialmente é o processo de maior ID

	// Cria um socket TCP para o processo ouvir os outros
	port = 4000 + eleicao_id
	addr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	fmt.Println("Socket aberta em ", port)
	go http.Serve(listener, nil)

	// Conecta aos sockets TCP dos outros processos
	for i := 0; i < numero_processos; i++ {
		if i != eleicao_id {
			port = 4000 + i
			addr, err = net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
			client, err := net.DialTCP("tcp", nil, addr)
			// Espera até a socket estar aberta para se conectar
			for err != nil {
				client, err = net.DialTCP("tcp", nil, addr)
			}
			clients = append(clients, client)
			fmt.Println("Conectado a socket em ", port)
		}
	}

	// Espera até a thread do servidor morrer
	wg.Wait()
}

func ReceiveMsg() {
	var msg string
	var splitMsg []string

	for {
		splitMsg = strings.Split(msg, "|")
		msgType := splitMsg[0]

		switch msgType {
		case "1":
			fmt.Println("Eleição")
		case "2":
			fmt.Println("OK")
		case "3":
			fmt.Println("Líder")
		case "4":
			fmt.Println("Vivo")
		case "5":
			fmt.Println("Vivo_OK")
		}
	}
}
