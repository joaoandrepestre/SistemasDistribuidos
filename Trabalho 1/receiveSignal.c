#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>


// Definição dos signal handlers

// Função chamada ao receber o sinal 14
static void handle14(int signal){
    printf("Recebido sinal 14.\n");
}

// Função chamada ao receber o sinal 54
static void handle54(int signal){
    printf("Recebido sinal 54.\n");
}

// Função chamada ao receber o sinal 60
// Termina o processo
static void handle60(int signal){
    printf("Recebido sinal 60.\nTerminando processo...\n");
    exit(0);
}

int main(int argc, char **argv){

    int pid, waitType;

    // Recupera PID do processo atual
    pid = getpid();

    // Instala signal handlers
    signal(14, handle14);
    signal(54, handle54);
    signal(60, handle60);

    // Checa se os parâmetros esperados foram passados
    if(argc<2){
        fprintf(stderr,"Forneça o tipo de wait esperado.\nbusy: 0 ou blocking: 1.\n");
        exit(1);
    }

    // Recupera os parâmetros passados
    waitType = atoi(argv[1]);

    // Verifica se os parâmetros passados estão adequados
    switch(waitType){
        case 0: 
            // busy
            printf("Tipo de wait: busy.\n");
            printf("Processo %d esperando sinais 14, 54 ou 60.\n", pid);
            
            // Espera infinita ocupando a CPU
            while(1);
            break;
        
        case 1: 
            // blocking
            printf("Tipo de wait: blocking.\n");
            printf("Processo %d esperando sinais 14, 54 ou 60.\n", pid);

            // Espera até a chegada de um sinal para ocupar a CPU
            while(1){
                pause();
            }
            break;
        default: // inexistente
            fprintf(stderr,"Tipo de wait desconhecido.\nEscolha entre busy: 0 ou blocking: 1.\n");
            exit(1);
    }

    return 0;
}