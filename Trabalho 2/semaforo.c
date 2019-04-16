#include <stdio.h>
#include <stdint.h>
#include <semaphore.h>
#include <pthread.h>
#include <stdlib.h>
#include <time.h>

#define BUFFER_SIZE 10
#define MAX_MEMORIA_CONSUMIDA 100000

void* produtor();
void* consumidor();
int consumir_numero_buffer();
void produzir_numero_buffer();
unsigned int procurar_espaco_livre();
unsigned int procurar_espaco_cheio();
void execucao_threads();
int checar_numero_primo();

int* numeros;

int numeros_produzidos;
int numeros_consumidos;

sem_t mutex;
sem_t vazio;
sem_t cheio;

int main (int argc, char** argv){

    srand(time(NULL));

    if (argc < 3){
        exit(1);
    }

    int num_threads_produtor = atoi(argv[1]);
    int num_threads_consumidor =  atoi(argv[2]);

    numeros = (int*) calloc(BUFFER_SIZE, sizeof(int));

    sem_init(&mutex, 0, 1);
    sem_init(&vazio, 0, BUFFER_SIZE);
    sem_init(&cheio, 0, 0);

    execucao_threads(num_threads_produtor, num_threads_consumidor);

    sem_destroy(&mutex);
    sem_destroy(&vazio);
    sem_destroy(&cheio);
}

void* produtor(){
    int numero_gerado;
    
    while(numeros_produzidos<MAX_MEMORIA_CONSUMIDA) {

        numero_gerado = rand()%100000;

        sem_wait(&vazio);
        sem_wait(&mutex);

        produzir_numero_buffer(numero_gerado);

        sem_post(&mutex);
        sem_post(&cheio);
    }
}

void* consumidor(){
    int numero_para_checar;

    while(numeros_consumidos<MAX_MEMORIA_CONSUMIDA) {
        sem_wait(&cheio);
        sem_wait(&mutex);
           
        numero_para_checar = consumir_numero_buffer();
     
        sem_post(&mutex);
        sem_post(&vazio);

        if(numero_para_checar != -1){

            int bool_primo = checar_numero_primo(numero_para_checar);

            if(bool_primo){
                printf("O número %ld é primo\n", numero_para_checar);
            } else {
                printf("O número %ld não é primo\n", numero_para_checar);
            }
        }
    }
}

void produzir_numero_buffer(int numero_gerado){

    if(numeros_produzidos < MAX_MEMORIA_CONSUMIDA){   

        int espaco_livre = procurar_espaco_livre();

        printf(" inserindo numero %d\n", numero_gerado);

        numeros[espaco_livre] = numero_gerado;

        numeros_produzidos++;
    }
}


int consumir_numero_buffer(){

    int temp = -1;

    if(numeros_consumidos < MAX_MEMORIA_CONSUMIDA){

        int espaco_cheio = procurar_espaco_cheio();

        temp = numeros[espaco_cheio];

        numeros[espaco_cheio] = 0;

        numeros_consumidos++;
    
    }

    return temp;
}

unsigned int procurar_espaco_livre(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(numeros[i] == 0){
            return i;
        }
    }
}

unsigned int procurar_espaco_cheio(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(numeros[i] != 0){
            return i;
        }
    }
}

int checar_numero_primo(unsigned int numero){
    unsigned int i;

    if(numero == 0) return 0;

    for(i=2;i<numero;i++){
        if(numero%i == 0) return 0;
    }

    return 1;
}

void execucao_threads(int num_threads_produtor, int num_threads_consumidor){
    pthread_t threads_produtor[num_threads_produtor];
    pthread_t threads_consumidor[num_threads_consumidor];
    
    int i;
    for(i = 0; i < num_threads_produtor; i++){
        pthread_create(&threads_produtor[i], NULL, produtor, NULL);
    }

    for(i = 0; i < num_threads_consumidor; i++){
        pthread_create(&threads_consumidor[i], NULL, consumidor, NULL);
    }

    for(i = 0; i < num_threads_produtor; i++){
        pthread_join(threads_produtor[i], NULL);
    }

    for(i = 0; i < num_threads_consumidor; i++){
        pthread_join(threads_consumidor[i], NULL);        
    }

    printf("Produzidos: %d, Consumidos: %d\n", numeros_produzidos, numeros_consumidos);
}