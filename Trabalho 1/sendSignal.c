#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

int main(int argc, char **argv){

    int pid, sig, err;

    // Checa se os parâmetros esperados foram passados
    if(argc!=3){
        printf("Número de parâmetros incorreto. Forneça o PID e o Sinal desejado.\n");
        return -1;
    }

    // Recupera os parâmetros passados
    pid = atoi(argv[1]);
    sig = atoi(argv[2]);

    // Checa se o processo escolhido existe
    // Envio de sinal NULL ao processo falha se o processo não existe
    if(kill(pid, 0) != 0){
        printf("Processo %d não existe. Forneça o PID de um processo existente.\n", pid);
        return -1;
    }

    // Envia o sinal ao processo escolhido
    err = kill(pid, sig);

    // Caso ocorra um erro no envio
    if(err != 0){
        printf("Ocorreu um erro no envio do sinal.\n");
        return -1;
    }

    // Caso não ocorra
    printf("Sinal %d enviado ao processo %d.\n", sig, pid);
    return 0;
}