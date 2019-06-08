package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"bufio"
	"time"
)

var pid int
var eleicao_id int
var lider_id int
var numero_processos int
var listener *net.TCPListener
var servers []*net.TCPConn
var clients []*net.TCPConn
var lider_vivo int
var recebeu_ok int

func main() {

	var wg sync.WaitGroup
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
	fmt.Println("Servindo em ", port)
	// Aceita conexões e armazena no array servers
	wg.Add(1)
	go func(){
		for i:=1;i<numero_processos;i++{
			conn, _ := listener.AcceptTCP()
			servers = append(servers, conn)
			fmt.Println("Aceitou conexão")
		}
	}()

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
			fmt.Println("Solicitou conexão de ", port)
			clients = append(clients, client)
		}
	}

	wg.Add(2)
	go ReceiveMsg()
	go CheckLider()
	// Espera até a thread do servidor morrer
	wg.Wait()
}

func CheckLider(){
	for{
		if lider_id != eleicao_id{
			lider_vivo = 0

			lider_index := lider_id
			if lider_id > eleicao_id{
				lider_index--
			}
			// Manda mensagem para lider para checar se está vivo
			fmt.Fprintf(clients[lider_index], "4|%d|%d\n", pid, eleicao_id)
			fmt.Println("Checou lider")

			time.AfterFunc(1*time.Second, func(){
				if lider_vivo != 1{
					fmt.Println("Lider não respondeu")
					// Se lider está morto, manda eleição para todos os processos
					recebeu_ok = 0
					for _, client := range clients{
						fmt.Println("Enviou eleição")
						fmt.Fprintf(client, "1|%d|%d\n", pid, eleicao_id)
					}
				}
			})
		}		
		time.Sleep(2*time.Second)
	}
}

func ReceiveMsg() {
	var mensagem string
	var err error

	if eleicao_id == numero_processos - 1{
		for {
			for _, server := range servers{

				mensagem, err = bufio.NewReader(server).ReadString('\n')
				fmt.Println("Recebeu mensagem ", mensagem)
				if err != nil {
					log.Fatal("Read error: ", err)
				}
				splitMsg := strings.Split(mensagem, "|")
				tipo, _ := strconv.Atoi(splitMsg[0])
				//pid_mensagem, _ := strconv.Atoi(splitMsg[1])
				eleicao_id_mensagem, _ := strconv.Atoi(splitMsg[2])

				switch tipo {
				case 1:
					fmt.Println("Eleição")
					eleicao(eleicao_id_mensagem)
				case 2:
					fmt.Println("OK")
					recebeu_ok = 1
				case 3:
					fmt.Println("Líder")
					lider_id = eleicao_id_mensagem
				case 4:
					fmt.Println("Vivo")
					index := eleicao_id_mensagem
					if eleicao_id_mensagem > eleicao_id{
						index--
					}
					fmt.Fprintf(clients[index], "5|%d|%d\n", pid, eleicao_id)
				case 5:
					fmt.Println("Vivo_OK")
					lider_vivo = 1
				}
			}
		}
	}
}

func eleicao(eleicao_id_mensagem int){
	if eleicao_id > eleicao_id_mensagem {
		fmt.Fprintf(clients[eleicao_id_mensagem], "2|%d|%d\n", pid, eleicao_id)
		recebeu_ok = 0
		for _, client := range clients{
			fmt.Fprintf(client, "1|%d|%d\n", pid, eleicao_id)
		}

		go func(){
			time.AfterFunc(1*time.Second, func(){
			 	if recebeu_ok != 1{
					for _, client := range clients{
						fmt.Fprintf(client, "3|%d|%d\n", pid, eleicao_id)
					}
				}
			})
		}()
	}
}
