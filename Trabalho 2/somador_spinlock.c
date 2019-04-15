#define _GNU_SOURCE
#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <time.h>
#include <pthread.h>
#include <unistd.h>
#include <sched.h>
#include "./lock/lock.h"

#define BUFFER_MIN_SIZE 10000000
#define BUFFER_MAX_SIZE 1000000000
#define MIN_NUM_THREADS 1
#define MAX_NUM_THREADS 256


void* somador_thread(void* threadNum);

void defineNumeros(int buffer_size);

void realizaSomatorio(int buffer_size);

void stick_this_thread_to_core(int core_id, pthread_attr_t* attr_addr);

// Lock para impedir acesso simultâneo ao somatório global
Lock *lock;

// Somatório global
int64_t somatorio;

// Números que serão somados
int8_t *numeros;
int tamanho;
int resto;


cpu_set_t cpus;

// Número de threads
int num_threads;

int main(int argc, char **argv){

    // Inicializa gerador de números aleatórios
    time_t seed;
    srand((unsigned) time(&seed));

    // Checa parâmetros
    if(argc < 2){
        fprintf(stderr, "Números de parâmetros incorreto. Informe o número de threads a serem criadas.\n");
        exit(1);
    }

    // Recupera parâmetros
    num_threads = atoi(argv[1]);

    // Inicia variáveis globais

    // Inicia lock
    lock = criaLock();
    
    // Inicia números que serão somados
    defineNumeros(BUFFER_MAX_SIZE);

    // Inicia medidor de tempo
    clock_t start, end;
    double tempo_cpu;

    int buffer_size;
    for(buffer_size=BUFFER_MIN_SIZE;buffer_size<=BUFFER_MAX_SIZE;buffer_size*=10){
        printf("Somatórios de %d valores.\n", buffer_size);
        for(num_threads=MIN_NUM_THREADS;num_threads<=MAX_NUM_THREADS;num_threads*=2){
            start = clock();
            int i;
            for(i=0;i<10;i++){
                realizaSomatorio(buffer_size);
            }
            end = clock();
            tempo_cpu = ((double) (end-start)) / (CLOCKS_PER_SEC*10);
            printf("\tSomatório usando %d threads: %lld. Levou %2f segundos.\n",num_threads, somatorio, tempo_cpu);
        }
        printf("\n");
    }

    return 0;
}

void *somador_thread(void* threadNum){

    // Define o indice de início da thread a partir do seu threadNum
    long t_num = (long) threadNum;
    int inic = t_num*tamanho;
    int tam = tamanho;

    if(t_num >= num_threads - resto){
        tam++;
        inic += resto - (num_threads-t_num);
    }

    // Realiza a soma parcial
    int64_t soma = 0;
    int i;
    for(i=0;i<tam;i++){
        soma += numeros[inic+i];
    }

    // Adiciona a soma parcial ao somatório total
    acquire(lock);
    somatorio += soma; // Protegido pelo lock
    release(lock);
}

void defineNumeros(int buffer_size){

    // Inicia array de números a serem somados
    numeros = (int8_t*) malloc(buffer_size*sizeof(int8_t));
    int i;
    for(i=0;i<buffer_size;i++){
        numeros[i] = (int8_t) (100 - rand()%201);
    }
}

void realizaSomatorio(int buffer_size){
    
    // Inicia tamanho do array que cada thread será responsável
    tamanho = buffer_size / num_threads;
    resto = buffer_size % num_threads;

    // Inicia somatório em 0
    somatorio = 0;

    // Cria threads
    pthread_t threads[num_threads];

    // Inicia atributos de thread
    pthread_attr_t attr;
    pthread_attr_init(&attr);

    long i;
    for(i=0;i<num_threads;i++){

        // Aloca cada thread em uma CPU diferente
        CPU_ZERO(&cpus);
        CPU_SET(i%4, &cpus);
        pthread_attr_setaffinity_np(&attr, sizeof(cpu_set_t), &cpus);
        
        // Cria a thread i
        pthread_create(&threads[i], &attr, somador_thread, (void*) i);
    }
    
    // Espera pelo fim das threads
    for(i=0;i<num_threads;i++){
        pthread_join(threads[i], NULL);
    }
}