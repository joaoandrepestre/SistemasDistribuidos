package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ThreadSafeInt - interface para impedir condição de corrida em variáveis globais
type ThreadSafeInt struct {
	value int32
}

// Get - utiliza biblioteca atomic para serializar a leitura ao valor armazenado
func (t *ThreadSafeInt) Get() int {
	return int(atomic.LoadInt32(&t.value))
}

// Set - utiliza biblioteca atomic para serializar a escrita ao valor armazenado
func (t *ThreadSafeInt) Set(v int) {
	atomic.StoreInt32(&t.value, int32(v))
}

var pid ThreadSafeInt
var eleicaoID ThreadSafeInt
var liderID ThreadSafeInt
var numeroProcessos ThreadSafeInt
var servers []*net.TCPConn
var clients []*net.TCPConn
var okTimer *time.Timer
var liderTimer *time.Timer

func main() {

	var wg sync.WaitGroup
	var port int

	// Recupera os argumentos da linha de comando
	args := os.Args[1:]
	// Checa o número de parâmetros
	if len(args) < 2 {
		log.Fatal("Número de argumentos inválido. Forneça o número de processos e o ID do processo atual.")
	}
	tmp, _ := strconv.Atoi(args[0])
	numeroProcessos.Set(tmp)
	tmp, _ = strconv.Atoi(args[1])
	eleicaoID.Set(tmp)
	// Checa se ID passado está dentro dos limites possíveis
	if eleicaoID.Get() < 0 || eleicaoID.Get() >= numeroProcessos.Get() {
		log.Fatal("Id para eleição inválido. Forneça um ID entre 0 e ", numeroProcessos.Get()-1)
	}

	// Inicia outras variáveis globais
	pid.Set(os.Getpid())
	liderID.Set(numeroProcessos.Get() - 1) // Líder inicialmente é o processo de maior ID

	// Cria um socket TCP para o processo ouvir os outros
	port = 4000 + eleicaoID.Get()
	addr, _ := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	fmt.Println("Servindo em ", port)

	wg.Add(1)
	// Aceita conexões e armazena no array servers
	go func() {
		defer wg.Done()
		for i := 1; i < numeroProcessos.Get(); i++ {
			conn, _ := listener.AcceptTCP()
			servers = append(servers, conn)
			fmt.Println("Aceitou conexão")
		}
	}()

	// Conecta aos sockets TCP dos outros processos
	for i := 0; i < numeroProcessos.Get(); i++ {
		if i != eleicaoID.Get() {
			port = 4000 + i
			addr, _ = net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
			client, err := net.DialTCP("tcp", nil, addr)
			// Espera até a socket estar aberta para se conectar
			for err != nil {
				client, err = net.DialTCP("tcp", nil, addr)
			}
			fmt.Println("Solicitou conexão de ", port)
			clients = append(clients, client)
		}
	}
	// Espera todas as conexões serem aceitas
	wg.Wait()

	wg.Add(2)
	go CheckLider()
	go ReceiveMsg()
	wg.Wait()
}

// CheckLider - função que periodicamente checa se líder está ativo
func CheckLider() {
	for {
		if liderID.Get() != eleicaoID.Get() {
			liderIndex := clientIndex(liderID.Get())

			// Manda mensagem para lider para checar se está vivo
			fmt.Fprintf(clients[liderIndex], "4|%d|%d\n", pid.Get(), eleicaoID.Get())
			fmt.Println("Checou lider")

			liderTimer = time.AfterFunc(1*time.Second, func() {
				fmt.Println("Lider não respondeu")
				// Se lider está morto, manda eleição para todos os processos
				enviarEleicao()
			})
		}
		time.Sleep(5 * time.Second)
	}
}

// ReceiveMsg - função que recebe e trata as mensagens recebidas
func ReceiveMsg() {
	var mensagem string
	var err error

	if eleicaoID.Get() != numeroProcessos.Get()-1 {
		for {
			for i, server := range servers {
				fmt.Println("Lendo mensagem de: ", i)
				mensagem, err = bufio.NewReader(server).ReadString('\n')
				fmt.Println("Recebeu mensagem ", mensagem)
				if err != nil {
					log.Fatal("Read error: ", err)
				}
				splitMsg := strings.Split(mensagem, "|")
				tipo, _ := strconv.Atoi(splitMsg[0])
				//pid_mensagem, _ := strconv.Atoi(splitMsg[1])
				eleicaoIDMensagem, _ := strconv.Atoi(splitMsg[2])
				fmt.Println("Recebeu mensagem de: ", eleicaoIDMensagem)

				switch tipo {
				case 1:
					fmt.Println("Eleição")
					go tratarEleicao(eleicaoIDMensagem)
				case 2:
					fmt.Println("OK")
					okTimer.Stop()
				case 3:
					fmt.Println("Líder")
					liderID.Set(eleicaoIDMensagem)
				case 4:
					fmt.Println("Vivo")
					index := clientIndex(eleicaoIDMensagem)
					fmt.Fprintf(clients[index], "5|%d|%d\n", pid.Get(), eleicaoID.Get())
				case 5:
					fmt.Println("Vivo_OK")
					liderTimer.Stop()
				}
			}
		}
	}
}

func tratarEleicao(eleicaoIDMensagem int) {
	if eleicaoID.Get() > eleicaoIDMensagem {
		fmt.Println("Meu ID é maior: ", eleicaoID.Get(), " > ", eleicaoIDMensagem)
		fmt.Fprintf(clients[eleicaoIDMensagem], "2|%d|%d\n", pid.Get(), eleicaoID.Get())
		fmt.Println("Enviei OK para: ", eleicaoIDMensagem)
		enviarEleicao()
	} else {
		fmt.Println("Meu ID é menor: ", eleicaoID.Get(), " < ", eleicaoIDMensagem)
	}
}

func enviarEleicao() {
	for _, client := range clients {
		fmt.Fprintf(client, "1|%d|%d\n", pid.Get(), eleicaoID.Get())
	}
	fmt.Println("Enviei ELEIÇÃO para todos")

	fmt.Println("Aguardando OK...")
	okTimer = time.AfterFunc(1*time.Second, func() {
		fmt.Println("Não recebi OK.")
		for _, client := range clients {
			fmt.Fprintf(client, "3|%d|%d\n", pid.Get(), eleicaoID.Get())
		}
		fmt.Println("Enviei LIDER para todos")
	})
}

func clientIndex(id int) int {
	index := id
	if id > eleicaoID.Get() {
		index--
	}
	return index
}
