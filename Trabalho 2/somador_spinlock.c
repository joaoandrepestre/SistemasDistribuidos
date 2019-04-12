#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <time.h>
#include <pthread.h>
#include "./lock/lock.h"

void *somador_thread(void* threadNum);

void defineNumeros(int N, int K);

// Lock para impedir acesso simultâneo ao somatório global
Lock *lock;

// Somatório global
int somatorio;

// Números que serão somados
int* numeros;
int tamanho;
int resto;


// Número de threads
int K;

int main(int argc, char **argv){

    // Inicializa gerador de números aleatórios
    time_t seed;
    srand((unsigned long) &seed);

    // Checa parâmetros
    if(argc < 2){
        fprintf(stderr, "Números de parâmetros incorreto. Informe o número de threads a serem criadas.\n");
        exit(1);
    }

    // Recupera parâmetros
    K = atoi(argv[1]);

    // Inicia variáveis globais

    // Inicia lock
    lock = criaLock();
    
    // Inicia somatório em 0
    somatorio = 0;

    // Inicia números que serão somados
    defineNumeros(1000000000, K);

    printf("Somatório antes: %d\n\n", somatorio);
    
    // Cria threads
    pthread_t threads[K];
    long i;
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
    long t_num = (long) threadNum;
    int inic = t_num*tamanho;
    int tam = tamanho;

    if(t_num >= K - resto){
        printf("%d\n",t_num);
        tam++;
        inic += resto - (K-t_num);
    }

    //printf("Args:\n\tinicio: %d\n\ttam:%d\n", inic, tam);

    // Realiza a soma parcial
    int soma = 0;
    int i;
    //printf("Numeros: ");
    for(i=0;i<tam;i++){
        //printf("%d; ", numeros[inic+i]);
        soma += numeros[inic+i];
    }
    //printf("\n");

    //printf("Soma parcial: %d\n\n", soma);

    // Adiciona a soma parcial ao somatório total
    acquire(lock);
    somatorio += soma; // Protegido pelo lock
    release(lock);
}

void defineNumeros(int N, int K){

    // Inicia array de números a serem somados
    numeros = (int*) malloc(N*sizeof(int));
    int i;
    for(i=0;i<N;i++){
        numeros[i] = -100 + rand()%200;
    }
    // Inicia tamanho do array que cada thread será responsável
    tamanho = N / K;
    resto = N % K;
}