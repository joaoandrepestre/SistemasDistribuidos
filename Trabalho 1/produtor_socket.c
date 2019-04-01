#include <unistd.h> 
#include <stdio.h> 
#include <sys/socket.h> 
#include <stdlib.h> 
#include <netinet/in.h> 
#include <string.h>
#include <time.h>

#define PORT 8080

int main(int argc, char const *argv[]) { 
    int server_fd, new_socket, valread; 
    struct sockaddr_in address; 
    int opt = 1; 
    int addrlen = sizeof(address); 
    char buffer[1024] = {0};  

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

	esperar_conexao: 
    	valread = read(new_socket, buffer, 1024);
		printf("%s\n", buffer);

	int i;
    for(i = 0; i < quantidade_num_gerados; i++){

        // Gera número aleatório crescente
        numero_aleatorio += rand()%100;
        printf("Número %d gerado.\n", numero_aleatorio);

        // Define mensagem a ser enviada
        sprintf(str_num_gerado, "%d", numero_aleatorio);

	    send(new_socket, str_num_gerado , strlen(str_num_gerado) , 0);

		valread = read(new_socket, buffer, 2);

		if(strcmp(buffer, "1") == 0) {
			printf("Número %d é primo\n", numero_aleatorio);
		} else {
			printf("Número %d não é primo\n", numero_aleatorio);
		}
    }    

	send(new_socket , "0", 1 , 0);

    printf("Quantidade máxima de números atingida. Terminando conexão com consumidor.\n");

	//goto esperar_conexao;
    
	return 0; 
} 