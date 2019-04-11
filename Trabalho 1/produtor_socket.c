#include <unistd.h> 
#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h>
#include <time.h>
#include<arpa/inet.h>

#define CONEXAO_ESTABELECIDA "1"
#define BUFFER_SIZE 20
#define TERMINATE "0"
#define PRIMO "1"
#define NAO_PRIMO "0"
#define PORT 8080

int main(int argc, char const *argv[]) { 
    int quantidade_num_gerados;
    char str_num_gerado[BUFFER_SIZE];

	// Inicia gerador de números aleatórios
    time_t seed;
    srand((unsigned)time(&seed));
	int numero_aleatorio = 1;
	
	if(argc < 2){
		fprintf(stderr, "Favor definir a quantidade de números gerados por conexão\n");
		exit(1);
	}

	quantidade_num_gerados = atoi(argv[1]);

    /*
     *  Início do código de conexão com sockets
     *  Retirado de: https://www.geeksforgeeks.org/socket-programming-cc/
     */

    struct sockaddr_in address; 
    int sock = 0, valread; 
    struct sockaddr_in serv_addr; 
    char *hello = CONEXAO_ESTABELECIDA; 
    char buffer[BUFFER_SIZE] = {0};

    if ((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) { 
        printf("\n Socket creation error \n"); 
        return -1; 
    } 
   
    memset(&serv_addr, '0', sizeof(serv_addr)); 
   
    serv_addr.sin_family = AF_INET; 
    serv_addr.sin_port = htons(PORT); 
       
    // Convert IPv4 and IPv6 addresses from text to binary form 
    if(inet_pton(AF_INET, "127.0.0.1", &serv_addr.sin_addr)<=0) { 
        printf("\nInvalid address/ Address not supported \n"); 
        return -1; 
    } 
   
    if (connect(sock, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) < 0) { 
        printf("\nConnection Failed \n"); 
        return -1; 
    } 

    /*
     *  Fim do código de conexão com sockets
     *  Retirado de: https://www.geeksforgeeks.org/socket-programming-cc/
     */

    // Função de escrita no canal de comunicação entre os sockets
    // O conteúdo da mensagem nesse caso não importa pois a mensagem é enviada apenas para
    // a confirmação da conexão e sincronização inicial entre os processos a partir de chamadas
    // bloqueantes
    send(sock , CONEXAO_ESTABELECIDA, BUFFER_SIZE, 0);

    // Loop de geração de números
	int i;
    for(i = 0; i < quantidade_num_gerados; i++) {
        // Gera número aleatório crescente
        numero_aleatorio += rand()%100;
        printf("Número %d gerado\n", numero_aleatorio);

        // Define mensagem a ser enviada
        sprintf(str_num_gerado, "%d", numero_aleatorio);

        // Envia o número para o consumidor
	    send(sock, str_num_gerado, BUFFER_SIZE , 0);

        // Lê a resposta do consumidor (tamanho da resposta é de 1 byte mais o string terminator \0)
		read(sock, buffer, 2);

		if(strcmp(buffer, PRIMO) == 0) {
			printf("Número %d é primo\n", numero_aleatorio);
		} else {
			printf("Número %d não é primo\n", numero_aleatorio);
		}
    }    

    // Envia ao consumidor o pedido para fechar a conexão (tamanho da requisição é de 1 byte mais o string terminator \0)
	send(sock, TERMINATE, 2 , 0);

    printf("Quantidade máxima de números atingida\nEnviando mensagem ao consumidor para fechar a conexão\n");
    
	return 0; 
} 