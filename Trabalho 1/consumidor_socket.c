#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h> 

#define TERMINATE "0"
#define BUFFER_SIZE 20
#define PORT 8080 

int numeroPrimo(int num) {
    int i;
    
    for(i = 2;i < num; i++) {
        if(num % i == 0) return 0;
    }

    return 1;
}

int main(int argc, char const *argv[]) { 
    char resposta[1];
    int num_analisado;

    /*
     *  Início do código de conexão com sockets
     *  Retirado de: https://www.geeksforgeeks.org/socket-programming-cc/
     */

    int server_fd, new_socket, valread; 
    struct sockaddr_in address; 
    int opt = 1; 
    int addrlen = sizeof(address); 
    char buffer[BUFFER_SIZE] = {0};  

	// Creating socket file descriptor 
    if ((server_fd = socket(AF_INET, SOCK_STREAM, 0)) == 0)  { 
        perror("socket failed"); 
        exit(EXIT_FAILURE); 
    } 
       
    // Forcefully attaching socket to the port 8080 
    if (setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR | SO_REUSEPORT, &opt, sizeof(opt))) { 
        perror("setsockopt"); 
        exit(EXIT_FAILURE); 
    }

    address.sin_family = AF_INET; 
    address.sin_addr.s_addr = INADDR_ANY; 
    address.sin_port = htons( PORT ); 
       
    // Forcefully attaching socket to the port 8080 
    if (bind(server_fd, (struct sockaddr *)&address, sizeof(address))<0) { 
        perror("bind failed"); 
        exit(EXIT_FAILURE); 
    } 

    if (listen(server_fd, 3) < 0) { 
        perror("listen"); 
        exit(EXIT_FAILURE); 
    }

    if ((new_socket = accept(server_fd, (struct sockaddr *)&address,(socklen_t*)&addrlen))<0) { 
        perror("accept"); 
        exit(EXIT_FAILURE); 
    }

    /*
     *  Fim do código de conexão com sockets
     *  Retirado de: https://www.geeksforgeeks.org/socket-programming-cc/
     */ 

    // Read bloqueante que recebe uma mensagem qualquer do produtor
    // Serve apenas para realizar uma sinconização inicial entre os dois programas
   	read(new_socket, buffer, BUFFER_SIZE);
    
    // Loop que lê o buffer enquanto o file descriptor existir ou até ele receber um sinal TERMINATE
    while(read(new_socket , buffer, BUFFER_SIZE) != -1 && strcmp(buffer, TERMINATE)) {
        printf("Número recebido %s\n", buffer);
        
        // Função que transforma string em um número da base 10
        num_analisado = strtol(buffer, NULL, 10);

        // Função que imprime em um buffer auxiliar de tamanho reduzido (1 byte) a resposta do servidor
        sprintf(resposta, "%d", numeroPrimo(num_analisado));

        // Função de envio da resposta ao produtor
        send(new_socket, resposta, 2, 0);
    }

    printf("Mensagem TERMINATE recebida do produtor\nFechando conexão\n");
    
    return 0; 
} 