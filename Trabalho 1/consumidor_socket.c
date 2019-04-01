#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h> 
#include<arpa/inet.h>

#define PORT 8080 
   
int numeroPrimo(int num){
    int i;
    
    for(i = 2;i < num; i++) {
        if(num % i == 0) return 0;
    }

    return 1;
}

int main(int argc, char const *argv[]) 
{ 
    struct sockaddr_in address; 
    int sock = 0, valread; 
    struct sockaddr_in serv_addr; 
    char *hello = "Conectado ao consumidor"; 
    char buffer[1024] = {0};

    char resposta[1];
    int num_analisado;
    
    if ((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) { 
        printf("\n Socket creation error \n"); 
        return -1; 
    } 
   
    memset(&serv_addr, '0', sizeof(serv_addr)); 
   
    serv_addr.sin_family = AF_INET; 
    serv_addr.sin_port = htons(PORT); 
       
    // Convert IPv4 and IPv6 addresses from text to binary form 
    if(inet_pton(AF_INET, "127.0.0.1", &serv_addr.sin_addr)<=0)  
    { 
        printf("\nInvalid address/ Address not supported \n"); 
        return -1; 
    } 
   
    if (connect(sock, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) < 0) 
    { 
        printf("\nConnection Failed \n"); 
        return -1; 
    } 

    send(sock , hello , strlen(hello) , 0 ); 
    
    while(/* strcmp(buffer, "0\0") &&  */read(sock , buffer, 1024) != -1) {
        printf("Número recebido %s\n", buffer);
        
        num_analisado = strtol(buffer, NULL, 10);

        sprintf(resposta, "%d", numeroPrimo(num_analisado));

        send(sock, resposta, 2, 0);
    }

    printf("Fechando conexão com o produtor\n");
    
    return 0; 
} 