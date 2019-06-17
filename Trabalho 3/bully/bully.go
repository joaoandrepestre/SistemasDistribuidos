package main

import (
	"io"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"../safe"
	"github.com/eiannone/keyboard"
	. "github.com/logrusorgru/aurora"
)

const msgELEICAO int = 1
const msgOK int = 2
const msgLIDER int = 3
const msgVIVO int = 4
const msgVIVOOK int = 5
const msgMORTO int = 6
const msgNAOOK int = 7


var pid safe.ThreadSafeInt
var eleicaoID safe.ThreadSafeInt
var liderID safe.ThreadSafeInt
var numeroProcessos safe.ThreadSafeInt
var conexoesLeitura []*net.TCPConn
var conexoesEscrita []*net.TCPConn

var recebiOK = safe.ThreadSafeBool{Value: false}
var recebiVIVOOK = safe.ThreadSafeBool{Value: false}

var contadorMensagem = safe.ThreadSafeInt{Value: 0}
var contadorDuranteUltimaEleicao = safe.ThreadSafeInt{Value: 0}
var contadorDuranteUltimaChecagemLider = safe.ThreadSafeInt{Value: 0}

var contadorMensagensRecebidas = [7]safe.ThreadSafeInt{{Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}}
var contadorMensagensEnviadas = [7]safe.ThreadSafeInt{{Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}, {Value: 0}}

var noMorto = safe.ThreadSafeBool{Value: false}
var checarLider = safe.ThreadSafeBool{Value: false}

var wg sync.WaitGroup

func main() {

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

	imprimirHelp()

	wg.Add(3)
	go CheckLider()
	go ReceiveMsg()
	go InterfaceTeclado()
	wg.Wait()
}

func imprimirHelp() {
	fmt.Printf("\nComandos da interface:\n" +
		"Ctrl l - Ativa/desativa checagem de líder (padrão: inativo)\n" +
		"Ctrl e - Imprime estatísticas de mensagem do nó\n" +
		"Ctrl m - Mata/revive nó (padrão: vivo)\n" +
		"Ctrl h - Imprimir esta lista de comandos\n" +
		"ESC - Finaliza o processo (pode causar READ ERROR nos demais processos)\n\n")
}

// InterfaceTeclado - thread que trata input de teclado
func InterfaceTeclado() {

	defer wg.Done()
	// c - checarLider
	// e - estatíticas
	// m - matar nó/reviver nó

	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		_, key, _ := keyboard.GetKey()

		switch key {
		case keyboard.KeyCtrlL:
			checarLider.Toggle()
			fmt.Println("\nChecar líder alterado:", checarLider.Get(), "\n")
		case keyboard.KeyCtrlE:
			imprimirEstatisticas()
		case keyboard.KeyCtrlM:
			noMorto.Toggle()
			fmt.Println("\nNó morto alterado:", noMorto.Get(), "\n")
		case keyboard.KeyCtrlH:
			imprimirHelp()
		case keyboard.KeyEsc:
			finalizarProcesso()
		case keyboard.KeyCtrlC:
			finalizarProcesso()
		}
	}
}

// CheckLider - função que periodicamente checa se líder está ativo
func CheckLider() {

	defer wg.Done()

	for {
		if liderID.Get() != eleicaoID.Get() && checarLider.Get() && liderID.Get() != -1 && !noMorto.Get() {
			contadorDuranteUltimaChecagemLider.Set(contadorMensagem.IncrementAndGet())
			recebiVIVOOK.Set(false)

			// Manda mensagem para lider para checar se está vivo
			enviarMensagemPara(msgVIVO, liderID.Get())
			fmt.Println(Blue("Mensagem"), Blue(contadorDuranteUltimaChecagemLider.Get()), Blue("- Enviei VIVO para:"), Blue(liderID.Get()))

			time.Sleep(3 * time.Second)

			if !recebiVIVOOK.Get() {
				fmt.Println(BrightRed("Mensagem"), BrightRed(contadorDuranteUltimaChecagemLider.Get()), BrightRed("- Lider não respondeu"))

				// Se lider está morto, manda eleição para todos os processos
				liderID.Set(-1)
				broadcastEleicao(true)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

// ReceiveMsg - função que recebe e trata as mensagens recebidas
func ReceiveMsg() {

	defer wg.Done()

	var mensagem string
	var err error
	var lock sync.Mutex

	for {
		for i, conn := range conexoesLeitura {
			go func(i int, conn *net.TCPConn) {
				var numeroDaMensagem int

				mensagem, err = bufio.NewReader(conn).ReadString('\n')
				numeroDaMensagem = contadorMensagem.IncrementAndGet()
				if err == io.EOF{
					lock.Lock()
					fmt.Println("Não foi possível ler de processo finalizado.")
					finalizarProcesso()
					lock.Unlock()
				} else if err != nil {
					log.Fatal("", err)
				}

				mensagemTratada := strings.Split(mensagem, "\n")
				splitMsg := strings.Split(mensagemTratada[0], "|")
				tipo, _ := strconv.Atoi(splitMsg[0])
				eleicaoIDMensagem, _ := strconv.Atoi(splitMsg[2])

				contadorMensagensRecebidas[tipo-1].Increment()

				if !noMorto.Get() {
					switch tipo {

					case msgELEICAO:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Eleição de"), Green(eleicaoIDMensagem))
						go tratarEleicao(eleicaoIDMensagem, numeroDaMensagem)

					case msgOK:
						contadorMensagem.Decrement()
						fmt.Println(Green("Mensagem"), Green(contadorDuranteUltimaEleicao.Get()), Green("- Recebi OK de"), Green(eleicaoIDMensagem))
						recebiOK.Set(true)

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
						recebiVIVOOK.Set(true)

					case msgMORTO:
						fmt.Println(Green("Mensagem"), Green(numeroDaMensagem), Green("- Recebi Morto de"), Green(eleicaoIDMensagem))

					case msgNAOOK:
						contadorMensagem.Decrement()
						fmt.Println(Green("Mensagem"), Green(contadorDuranteUltimaEleicao.Get()), Green("- Recebi Não_OK de"), Green(eleicaoIDMensagem))

					}
				} else {
					if tipo == msgLIDER {
						liderID.Set(eleicaoIDMensagem)
					}
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

	recebiOK.Set(false)

	for _, conn := range conexoesEscrita {
		contadorMensagensEnviadas[msgELEICAO-1].Increment()
		fmt.Fprintf(conn, "1|%d|%d\n", pid.Get(), eleicaoID.Get())
	}

	fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Enviei ELEIÇÃO para todos"))
	fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Aguardando OK..."))

	time.Sleep(3 * time.Second)

	if !recebiOK.Get() {
		fmt.Println(BrightRed("Mensagem"), BrightRed(numeracaoContador), BrightRed("- Não recebi OK."))
		broadcastLider()
		fmt.Println(Blue("Mensagem"), Blue(numeracaoContador), Blue("- Enviei LIDER para todos"))
	}
}

func broadcastLider() {
	liderID.Set(eleicaoID.Get())
	for _, conn := range conexoesEscrita {
		contadorMensagensEnviadas[msgLIDER-1].Increment()
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

func imprimirEstatisticas() {
	fmt.Printf("\nMENSAGENS:\n"+
		"Eleição: 	%d enviadas, %d recebidas\n"+
		"OK: 		%d enviadas, %d recebidas\n"+
		"Líder:		%d enviadas, %d recebidas\n"+
		"Vivo: 		%d enviadas, %d recebidas\n"+
		"Vivo_OK: 	%d enviadas, %d recebidas\n"+
		"Morto:		%d enviadas, %d recebidas\n"+
		"Não_OK: 	%d enviadas, %d recebidas\n\n"+
		"Processo Líder: %d\n\n",
		contadorMensagensEnviadas[msgELEICAO-1].Get(), contadorMensagensRecebidas[msgELEICAO-1].Get(),
		contadorMensagensEnviadas[msgOK-1].Get(), contadorMensagensRecebidas[msgOK-1].Get(),
		contadorMensagensEnviadas[msgLIDER-1].Get(), contadorMensagensRecebidas[msgLIDER-1].Get(),
		contadorMensagensEnviadas[msgVIVO-1].Get(), contadorMensagensRecebidas[msgVIVO-1].Get(),
		contadorMensagensEnviadas[msgVIVOOK-1].Get(), contadorMensagensRecebidas[msgVIVOOK-1].Get(),
		contadorMensagensEnviadas[msgMORTO-1].Get(), contadorMensagensRecebidas[msgMORTO-1].Get(),
		contadorMensagensEnviadas[msgNAOOK-1].Get(), contadorMensagensRecebidas[msgNAOOK-1].Get(),
		liderID.Get())
}

func finalizarProcesso(){
	fmt.Println("\nFinalizando processo...\n")
	keyboard.Close()
	os.Exit(0)
}
