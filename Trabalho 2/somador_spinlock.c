#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <pthread.h>
#include "./lock/lock.h"

void *somador_thread(void* threadNum);

Lock *lock;
int somatorio;

int* numeros;
int inicio;
int tamanho;

int main(int argc, char **argv){

    // Checa parâmetros
    if(argc < 2){
        fprintf(stderr, "Números de parâmetros incorreto. Informe o número de threads a serem criadas.\n");
        exit(1);
    }

    // Recupera parâmetros
    int K = atoi(argv[1]);

    // Inicia variáveis globais

    // Inicia lock
    lock = criaLock();
    
    // Inicia somatório em 0
    somatorio = 0;

    // Inicia array de números a serem somados
    numeros = (int*) malloc(4*sizeof(int));
    long i;
    for(i=0;i<4;i++){
        numeros[i] = i+1;
    }
    // Inicia tamanho do array que cada thread será responsável
    tamanho = 4/K;
    // Inicia inidice de partida da soma de cada thread como 0
    inicio = 0;
    

    printf("Somatório antes: %d\n\n", somatorio);
    
    // Cria threads
    pthread_t threads[K];
    for(i=0;i<K;i++){
        pthread_create(&threads[i], NULL, somador_thread, (void*) i);
    }
    
    // Espera pelo fim das threads
    for(i=0;i<K;i++){
        pthread_join(threads[i], NULL);
    }
    

    printf("Somatório: %d\n", somatorio);

    return 0;
}

void *somador_thread(void* threadNum){

    // Define o indice de início da thread a partir do seu threadNum
    int inic = inicio + (long)threadNum*tamanho;

    printf("Args:\n\tinicio: %d\n\ttam:%d\n", inic, tamanho);

    // Realiza a soma parcial
    int soma = 0;
    int i;
    printf("Numeros: ");
    for(i=0;i<tamanho;i++){
        printf("%d; ", numeros[inic+i]);
        soma += numeros[inic+i];
    }
    printf("\n");

    printf("Soma parcial: %d\n\n", soma);

    // Adiciona a soma parcial ao somatório total
    acquire(lock);
    somatorio += soma; // Protegido pelo lock
    release(lock);
}