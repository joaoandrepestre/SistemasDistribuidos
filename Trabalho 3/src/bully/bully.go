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

	. "github.com/logrusorgru/aurora"
)

const msgELEICAO int = 1
const msgOK int = 2
const msgLIDER int = 3
const msgVIVO int = 4
const msgVIVOOK int = 5
const msgMORTO int = 6
const msgNAOOK int = 7

const liderMorto int = -1

const TRUE int = 1
const FALSE int = 0

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

// IncrementAndGet - incrementa valor em 1 e retorna
func (t *ThreadSafeInt) IncrementAndGet() int {
	atomic.StoreInt32(&t.value, int32(t.Get()+1))
	return int(atomic.LoadInt32(&t.value))
}

// Decrement - decrementa o valor em 1
func (t *ThreadSafeInt) Decrement() {
	atomic.StoreInt32(&t.value, int32(t.Get()-1))
}

// Increment - incrementa o valor em 1
func (t *ThreadSafeInt) Increment() {
	atomic.StoreInt32(&t.value, int32(t.Get()+1))
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
var recebiOK = ThreadSafeInt{value: 0}
var recebiVIVOOK = ThreadSafeInt{value: 0}

var contadorMensagem = ThreadSafeInt{value: 0}
var contadorDuranteUltimaEleicao = ThreadSafeInt{value: 0}
var contadorDuranteUltimaChecagemLider = ThreadSafeInt{value: 0}

var contadorMensagensRecebidas = [7]ThreadSafeInt{{value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}}
var contadorMensagensEnviadas = [7]ThreadSafeInt{{value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}, {value: 0}}

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

	fmt.Println(Yellow("Servindo em "), Yellow(port))
	wg.Add(1)

	// Aceita conexões e armazena no array servers
	go func() {
		defer wg.Done()
		for i := 1; i < numeroProcessos.Get(); i++ {
			conn, _ := listener.AcceptTCP()
			conexoesLeitura = append(conexoesLeitura, conn)
			fmt.Println(Yellow("Aceitou conexão"))
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

			fmt.Println(Yellow("Solicitou conexão de "), Yellow(port))
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
		if liderID.Get() != eleicaoID.Get() && liderID.Get() != liderMorto {
			contadorDuranteUltimaChecagemLider.Set(contadorMensagem.IncrementAndGet())
			recebiVIVOOK.Set(FALSE)

			// Manda mensagem para lider para checar se está vivo
			// fmt.Println("Lider atual: ", liderID.Get())

			enviarMensagemPara(msgVIVO, liderID.Get())
			fmt.Println(Blue("Mensagem"), Blue(contadorDuranteUltimaChecagemLider.Get()), Blue("- Enviei VIVO para:"), Blue(liderID.Get()))
			// fmt.Println("Checou lider")

			time.Sleep(3 * time.Second)

			if recebiVIVOOK.Get() == FALSE {
				fmt.Println(BrightRed("Mensagem"), BrightRed(contadorDuranteUltimaChecagemLider.Get()), BrightRed("- Lider não respondeu"))

				// Se lider está morto, manda eleição para todos os processos
				liderID.Set(liderMorto)
				broadcastEleicao(true)
			}
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
				var numeroDaMensagem int

				mensagem, err = bufio.NewReader(conn).ReadString('\n')
				numeroDaMensagem = contadorMensagem.IncrementAndGet()

				if err != nil {
					log.Fatal("Read error: ", err)
				}

				mensagemTratada := strings.Split(mensagem, "\n")
				splitMsg := strings.Split(mensagemTratada[0], "|")
				tipo, _ := strconv.Atoi(splitMsg[0])
				eleicaoIDMensagem, _ := strconv.Atoi(splitMsg[2])

				contadorMensagensRecebidas[tipo-1].Increment()

				if eleicaoID.Get() != numeroProcessos.Get()-1 {
					switch tipo {

					case msgELEICAO:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Eleição de"), Green(eleicaoIDMensagem))
						go tratarEleicao(eleicaoIDMensagem, numeroDaMensagem)

					case msgOK:
						contadorMensagem.Decrement()
						fmt.Println(Green("Mensagem"), Green(contadorDuranteUltimaEleicao.Get()), Green("- Recebi OK de"), Green(eleicaoIDMensagem))
						recebiOK.Set(TRUE)

					case msgLIDER:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Líder de"), Green(eleicaoIDMensagem))
						liderID.Set(eleicaoIDMensagem)
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Novo líder: "), Green(liderID.Get()))

					case msgVIVO:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Vivo de"), Green(eleicaoIDMensagem))
						fmt.Println(Blue("Mensagem"), Blue(numeroDaMensagem), Blue("- Enviei VIVO_OK para:"), Blue(eleicaoIDMensagem))
						enviarMensagemPara(msgVIVOOK, eleicaoIDMensagem)

					case msgVIVOOK:
						contadorMensagem.Decrement()
						fmt.Println(Green("Mensagem"), Green(contadorDuranteUltimaChecagemLider.Get()), Green("- Recebi Vivo_OK de"), Green(eleicaoIDMensagem))
						recebiVIVOOK.Set(TRUE)

					case msgMORTO:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Morto de"), Green(eleicaoIDMensagem))

					case msgNAOOK:
						contadorMensagem.Decrement()
						fmt.Println(Green("Mensagem"), Green(contadorDuranteUltimaEleicao.Get()), Green("- Recebi Não_OK de"), Green(eleicaoIDMensagem))

					}
				} else {
					fmt.Println(Blue("Mensagem"), Blue(numeroDaMensagem), Blue("- Enviei MORTO para:"), Blue(eleicaoIDMensagem))
					enviarMensagemPara(msgMORTO, eleicaoIDMensagem)
				}
			}(i, conn)
		}
		time.Sleep(2 * time.Second)
	}
}

func tratarEleicao(eleicaoIDMensagem int, numeroDaMensagem int) {
	if eleicaoID.Get() > eleicaoIDMensagem {
		enviarMensagemPara(msgOK, eleicaoIDMensagem)
		fmt.Println(Blue("Mensagem"), Blue(numeroDaMensagem), Blue("- Enviei OK para:"), Blue(eleicaoIDMensagem))
		broadcastEleicao(false)
	} else {
		enviarMensagemPara(msgNAOOK, eleicaoIDMensagem)
		fmt.Println(Blue("Mensagem"), Blue(numeroDaMensagem), Blue("- Enviei Não_OK para:"), Blue(eleicaoIDMensagem))
	}
}

func enviarMensagemPara(tipoMensagem int, id int) {
	contadorMensagensEnviadas[tipoMensagem-1].Increment()
	index := clientIndex(id)
	fmt.Fprintf(conexoesEscrita[index], "%d|%d|%d\n", tipoMensagem, pid.Get(), eleicaoID.Get())
}

func broadcastEleicao(eleicaoPorLiderEstarMorto bool) {
	if eleicaoPorLiderEstarMorto {
		contadorDuranteUltimaEleicao.Set(contadorDuranteUltimaChecagemLider.Get())
	} else {
		contadorDuranteUltimaEleicao.Set(contadorMensagem.IncrementAndGet())
	}

	var numeracaoContador = contadorDuranteUltimaEleicao.Get()

	recebiOK.Set(FALSE)

	for _, conn := range conexoesEscrita {
		fmt.Fprintf(conn, "1|%d|%d\n", pid.Get(), eleicaoID.Get())
	}

	fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Enviei ELEIÇÃO para todos"))
	fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Aguardando OK..."))

	time.Sleep(3 * time.Second)

	if recebiOK.Get() == FALSE {
		fmt.Println(BrightRed("Mensagem"), BrightRed(numeracaoContador), BrightRed("- Não recebi OK."))
		broadcastLider()
		fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Enviei LIDER para todos"))
	}
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
