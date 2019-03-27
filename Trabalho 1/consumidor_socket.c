#include <unistd.h> 
#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <arpa/inet.h> 
#include <string.h> 

#define PORT 1414

int main(int argc, char const* argv[]){

    struct sockaddr_in addr;
    int socket_fd = 0;
    struct sockaddr_in prod_addr;

    if(socket_fd = socket(AF_INET, SOCK_STREAM, 0) < 0){
        fprintf(stderr, "Falha na criação do socket.\n");
        exit(1);
    }

    memset(&prod_addr, '0', sizeof(prod_addr));

    prod_addr.sin_family = AF_INET;
    prod_addr.sin_port = htons(PORT);

    if(inet_pton(AF_INET, "127.0.0.1", &prod_addr.sin_addr)<=0){
        fprintf(stderr, "Endereço inválido.\n");
        exit(1);
    }

    if(connect(socket_fd, (struct sockaddr *)&prod_addr, sizeof(prod_addr))<0){
        fprintf(stderr, "Falha na conexão.\n");
        exit(1);
    }

    send(socket_fd, "TESTE", 5, 0);
    return 0;
}