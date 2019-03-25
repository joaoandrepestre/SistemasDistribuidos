#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

int main(int argc, char **argv){

    int qt, fork_ret;

    // Define pipe
    int pip[2];
    pipe(pip);

    // Checa se os parâmetros esperados foram passados
    if(argc < 2){
        printf("Forneça a quantidade de números a serem gerados.\n");
        return -1;
    }

    // Recupera os parâmetros
    qt = atoi(argv[1]);

    // Divide o processo em pai e filho
    fork_ret = fork();

    // Processo pai(agindo como produtor)
    if(fork_ret > 0){
        // Produzir
    } else{
        // Consumir
    }



    return 0;
}