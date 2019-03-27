#include <unistd.h> 
#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h> 

#define PORT 1414

int main(int argc, char const* argv[]){

    // File descriptors para programa produtor e para socket
    int prod_fd, socket_fd;

    // Endereço do socket
    struct sockaddr_in addr;
    int addrlen = sizeof(addr);

    // Estabelece o file descriptor para produtor
    if ((prod_fd = socket(AF_INET, SOCK_STREAM, 0)) == 0 ){
        fprintf(stderr, "Falha na criação do socket.n");
        exit(1);
    }

    // Define protocolo como IPV4
    addr.sin_family = AF_INET; 

    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons( PORT ); 
       
    // Associa socket a porta 14
    if (bind(prod_fd, (struct sockaddr *)&addr, sizeof(addr))<0) { 
        fprintf(stderr,"Falha na associação à porta.\n"); 
        exit(1); 
    } 
    
    // Espera conexão com o consumidor
    if (listen(prod_fd, 3) < 0) { 
        fprintf(stderr,"Falha na espera por conexão.\n"); 
        exit(1); 
    }

    // Aceita pedido de conexão 
    if(socket_fd = accept(prod_fd, (struct sockaddr *)&addr, (socklen_t*)&addrlen)<0){
        fprintf(stderr, "Falha na conexão.\n");
        exit(1);
    }

    char leu[5];
    read(socket_fd, leu, 5);
    printf("Leu: %s\n", leu);

    return 0;
}