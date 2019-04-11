#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <pthread.h>
#include "./lock/lock.h"

typedef struct{
    int *numeros;
    int tam;
} thread_args;

void *somador_thread(thread_args* args);

Lock *lock;
int somatorio;

int main(int argc, char **argv){

    // Checa parâmetros
    if(argc < 2){
        fprintf(stderr, "Números de parâmetros incorreto. Informe o número de threads a serem criadas.\n");
        exit(1);
    }

    // Recupera parâmetros
    int K = atoi(argv[1]);

    // Inicia lock
    lock = criaLock();

    somatorio = 0;
    int numeros[5] = {1,1,1,1,1};
    thread_args args;
    args.numeros = numeros;
    args.tam = 5;

    printf("Somatório antes: %d\n", somatorio);
    // Cria threads
    pthread_t thread;
    pthread_create(&thread, NULL, somador_thread, &args);
    pthread_join(thread, NULL);

    printf("Somatório: %d\n", somatorio);

    return 0;
}

void *somador_thread(thread_args* args){

    int *numeros = args->numeros;
    int tam = args->tam;

    int soma = 0;
    int i;
    for(i=0;i<tam;i++){
        soma += numeros[i];
    }

    acquire(lock);
    somatorio += soma;
    release(lock);
}