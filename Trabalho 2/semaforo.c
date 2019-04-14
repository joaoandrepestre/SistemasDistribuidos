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
long int consumir_numero_buffer();
void produzir_numero_buffer();
unsigned int procurar_espaco_livre();
unsigned int procurar_espaco_cheio();
void execucao_threads();
int checar_numero_primo();

long int* numeros;

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

    numeros = (long*) calloc(BUFFER_SIZE*sizeof(long), 0);

    sem_init(&mutex, 0, 1);
    sem_init(&vazio, 0, BUFFER_SIZE);
    sem_init(&cheio, 0, 0);

    execucao_threads(num_threads_produtor, num_threads_consumidor);

    sem_destroy(&mutex);
    sem_destroy(&vazio);
    sem_destroy(&cheio);
}

void* produtor(){
    long int numero_gerado;
    
    while(1) {

        numero_gerado = rand();

        sem_wait(&vazio);
        sem_wait(&mutex);

        produzir_numero_buffer(numero_gerado);

        sem_post(&mutex);
        sem_post(&cheio);
    }
}

void* consumidor(){
    long int numero_para_checar;

    while(1) {
        sem_wait(&cheio);
        sem_wait(&mutex);
           
        numero_para_checar = consumir_numero_buffer();
     
        sem_post(&mutex);
        sem_post(&vazio);

        int bool_primo = checar_numero_primo(numero_para_checar);

        if(bool_primo){
            printf("O número %d é primo\n", numero_para_checar);
        } else {
            printf("O número %d não é primo\n", numero_para_checar);
        }
    }
}

void produzir_numero_buffer(long int numero_gerado){
    int espaco_livre = procurar_espaco_livre();

    numeros[espaco_livre] = numero_gerado;
}


long int consumir_numero_buffer(){
    int espaco_cheio = procurar_espaco_cheio();

    long int temp = numeros[espaco_cheio];

    numeros[espaco_cheio] = 0;

    return temp;
}

unsigned int procurar_espaco_livre(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(numeros[i] == 0){
            printf("espaço vazio %d\n", i);
            return i;
        }
    }
    printf("coe\n");

}

unsigned int procurar_espaco_cheio(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(numeros[i] != 0){
            printf("espaço cheio %d\n", i);
            return i;
        }
    }
    printf("coeeeeeez\n");
}

int checar_numero_primo(unsigned int numero){
    unsigned int i;
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
}