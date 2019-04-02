#include <unistd.h> 
#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h>
#include <time.h>
#include<arpa/inet.h>


#define PORT 8080

int main(int argc, char const *argv[]) { 

    int quantidade_num_gerados;
    char str_num_gerado[20];

	// Inicia gerador de números aleatórios
    time_t seed;
    srand((unsigned)time(&seed));
	int numero_aleatorio = 1;
	
	if(argc < 2){
		fprintf(stderr, "Favor definir a quantidade de números gerados por conexão\n");
		exit(1);
	}

	quantidade_num_gerados = atoi(argv[1]);

    /* Início do código de conexão com sockets
    retirado de: https://www.geeksforgeeks.org/socket-programming-cc/ */

    struct sockaddr_in address; 
    int sock = 0, valread; 
    struct sockaddr_in serv_addr; 
    char *hello = "Conectado ao produtor"; 
    char buffer[1024] = {0};

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

    /* Fim do código de conexão com sockets
    retirado de: https://www.geeksforgeeks.org/socket-programming-cc/ */

	int i;
    for(i = 0; i < quantidade_num_gerados; i++){

        // Gera número aleatório crescente
        numero_aleatorio += rand()%100;
        printf("Número %d gerado.\n", numero_aleatorio);

        // Define mensagem a ser enviada
        sprintf(str_num_gerado, "%d", numero_aleatorio);

	    send(sock, str_num_gerado , strlen(str_num_gerado)+1 , 0);

		valread = read(sock, buffer, 2);

		if(strcmp(buffer, "1") == 0) {
			printf("Número %d é primo\n", numero_aleatorio);
		} else {
			printf("Número %d não é primo\n", numero_aleatorio);
		}
    }    

	send(sock , "0", 2 , 0);

    printf("Quantidade máxima de números atingida. Terminando conexão com consumidor.\n");
    
	return 0; 
} 