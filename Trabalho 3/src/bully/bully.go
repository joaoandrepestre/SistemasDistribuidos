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

const msgELEICAO int = 1
const msgOK int = 2
const msgLIDER int = 3
const msgVIVO int = 4
const msgVIVOOK int = 5
const msgMORTO int = 6
const msgNAOOK int = 7

const liderMorto int = -1

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

/* // ThreadSafeTimer - interface para impedir condição de corrida em variáveis globais
type ThreadSafeTimer struct {
	timer *time.Timer
	mutex sync.Mutex
}

// Set - utiliza mutex para serializar escrita no timer
func (t *ThreadSafeTimer) Set(timer *time.Timer) {
	t.mutex.Lock()
	if t.timer == nil {
		t.timer = timer
	} else {
		t.timer.Reset(5 * time.Second)
	}
	t.mutex.Unlock()
} */

var pid ThreadSafeInt
var eleicaoID ThreadSafeInt
var liderID ThreadSafeInt
var numeroProcessos ThreadSafeInt
var conexoesLeitura []*net.TCPConn
var conexoesEscrita []*net.TCPConn
var okChan chan int
var liderChan chan int

/* var okTimer *time.Timer
var liderTimer *time.Timer */

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
			conexoesLeitura = append(conexoesLeitura, conn)
			fmt.Println("Aceitou conexão")
		}
	}()

	// Conecta aos sockets TCP dos outros processos
	for i := 0; i < numeroProcessos.Get(); i++ {
		if i != eleicaoID.Get() {
			port = 4000 + i
			addr, _ = net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
			conn, err := net.DialTCP("tcp", nil, addr)
			// Espera até a socket estar aberta para se conectar
			for err != nil {
				conn, err = net.DialTCP("tcp", nil, addr)
			}
			fmt.Println("Solicitou conexão de ", port)
			conexoesEscrita = append(conexoesEscrita, conn)
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
		if eleicaoID.Get() == 0 && liderID.Get() != eleicaoID.Get() && liderID.Get() != liderMorto {

			// Manda mensagem para lider para checar se está vivo
			fmt.Println("Lider atual: ", liderID.Get())
			enviarMensagemPara(msgVIVO, liderID.Get())
			fmt.Println("Checou lider")

			liderVivo := <-liderChan
			if liderVivo != msgVIVOOK {
				fmt.Println("Lider não respondeu")
				// Se lider está morto, manda eleição para todos os processos
				liderID.Set(liderMorto)
				broadcastEleicao()
			}
			/* liderTimer = time.AfterFunc(5*time.Second, func() {
				fmt.Println("Lider não respondeu")
				// Se lider está morto, manda eleição para todos os processos
				liderID.Set(-1)
				broadcastEleicao()
			}) */
		}
		time.Sleep(10 * time.Second)
	}
}

// ReceiveMsg - função que recebe e trata as mensagens recebidas
func ReceiveMsg() {
	var mensagem string
	var err error

	for {
		for i, conn := range conexoesLeitura {
			go func(i int, conn *net.TCPConn) {
				fmt.Println("Lendo mensagem de: ", i)
				mensagem, err = bufio.NewReader(conn).ReadString('\n')
				fmt.Println("Recebeu mensagem ", mensagem)
				if err != nil {
					log.Fatal("Read error: ", err)
				}
				splitMsg := strings.Split(mensagem[:len(mensagem)-1], "|")
				tipo, _ := strconv.Atoi(splitMsg[0])
				//pid_mensagem, _ := strconv.Atoi(splitMsg[1])
				eleicaoIDMensagem, _ := strconv.Atoi(splitMsg[2])
				fmt.Println("Recebeu mensagem de: ", eleicaoIDMensagem)

				if eleicaoID.Get() != numeroProcessos.Get()-1 {
					switch tipo {
					case msgELEICAO:
						fmt.Println("Eleição")
						go tratarEleicao(eleicaoIDMensagem)
					case msgOK:
						fmt.Println("OK")
						//okTimer.Stop()
						okChan <- msgOK
					case msgLIDER:
						fmt.Println("Líder")
						liderID.Set(eleicaoIDMensagem)
						fmt.Println("Novo líder: ", liderID.Get())
					case msgVIVO:
						fmt.Println("Vivo")
						enviarMensagemPara(msgVIVOOK, eleicaoIDMensagem)
					case msgVIVOOK:
						fmt.Println("Vivo_OK")
						//liderTimer.Stop()
						liderChan <- msgVIVOOK
					case msgMORTO:
						fmt.Println("Morto")
						liderChan <- msgMORTO
					case msgNAOOK:
						fmt.Println("Não_OK")
						okChan <- msgNAOOK
					}
				} else {
					enviarMensagemPara(msgMORTO, eleicaoIDMensagem)
				}
			}(i, conn)
		}
	}

}

func tratarEleicao(eleicaoIDMensagem int) {
	if eleicaoID.Get() > eleicaoIDMensagem {
		fmt.Println("Meu ID é maior: ", eleicaoID.Get(), " > ", eleicaoIDMensagem)
		enviarMensagemPara(msgOK, eleicaoIDMensagem)
		fmt.Println("Enviei OK para: ", eleicaoIDMensagem)
		broadcastEleicao()
	} else {
		fmt.Println("Meu ID é menor: ", eleicaoID.Get(), " < ", eleicaoIDMensagem)
	}
}

func enviarMensagemPara(tipoMensagem int, id int) {
	index := clientIndex(id)
	fmt.Fprintf(conexoesEscrita[index], "%d|%d|%d\n", tipoMensagem, pid.Get(), eleicaoID.Get())
}

func broadcastEleicao() {
	for _, conn := range conexoesEscrita {
		fmt.Fprintf(conn, "1|%d|%d\n", pid.Get(), eleicaoID.Get())
	}
	fmt.Println("Enviei ELEIÇÃO para todos")

	fmt.Println("Aguardando OK...")
	ok := <-okChan
	if ok != msgOK {
		fmt.Println("Não recebi OK.")
		broadcastLider()
		fmt.Println("Enviei LIDER para todos")
	}
	/* okTimer = time.AfterFunc(5*time.Second, func() {
		fmt.Println("Não recebi OK.")
		broadcastLider()
		fmt.Println("Enviei LIDER para todos")
	}) */
}

func broadcastLider() {
	liderID.Set(eleicaoID.Get())
	for _, conn := range conexoesEscrita {
		fmt.Fprintf(conn, "3|%d|%d\n", pid.Get(), eleicaoID.Get())
	}
}

func clientIndex(id int) int {
	index := id
	if id > eleicaoID.Get() {
		index--
	}
	return index
}
