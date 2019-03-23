#include <stdlib.h>
#include <stdio.h>

int main(int argc, char **argv){

    if(argc!=3){
        printf("Número de parâmetros incorreto. Forneça o PID e o Sinal desejado.\n");
        return -1;
    }

    int pid = atoi(argv[1]);
    int sig = atoi(argv[2]);
    printf("PID: %d\nSignal: %d\n",pid, sig);
    return 0;
}