#include <stdio.h>
#include <stdint.h>
#include <semaphore.h>
#include <pthread.h>
#include <stdlib.h>
#include <time.h>

// Define tamanho da memória compartilhada
#define BUFFER_SIZE 10

// Define máximo de números a serem trocados entre produtores e consumirdores
#define MAX_MEMORIA_CONSUMIDA 100000

void* produtor();
void* consumidor();
int consumir_numero_buffer();
void produzir_numero_buffer();
unsigned int procurar_espaco_livre();
unsigned int procurar_espaco_cheio();
void execucao_threads();
int checar_numero_primo();

// Buffer de memória compartilhada
int* buffer_compartilhado;

// Quantidade de números produzidos
int numeros_produzidos;

// Quantidade de números consumidos
int numeros_consumidos;

// Semáforos para sincronização das threads
sem_t mutex;
sem_t vazio;
sem_t cheio;

int main (int argc, char** argv){

    // Inicia gerador de números aleatórios
    srand(time(NULL));

    // Checa se o número de parâmetros está correto
    if (argc < 3){
        fprintf(stderr, "Selecione o número de threads produtoras e consumidoras.\n");
        exit(1);
    }

    // Recupera os parâmetros passados
    int num_threads_produtor = atoi(argv[1]);
    int num_threads_consumidor =  atoi(argv[2]);

    // Aloca o buffer de memória compartilhada iniciando ele vazio
    buffer_compartilhado = (int*) calloc(BUFFER_SIZE, sizeof(int));

    // Inicia os semáforos:
    // Semáforo mutex para acesso ao buffer
    sem_init(&mutex, 0, 1);
    // Semáforo contador de espaços desocupados
    sem_init(&vazio, 0, BUFFER_SIZE);
    // Semáforo contador de espaços ocupados
    sem_init(&cheio, 0, 0);

    // Executa as threads
    execucao_threads(num_threads_produtor, num_threads_consumidor);

    // Destrói os semáforos
    sem_destroy(&mutex);
    sem_destroy(&vazio);
    sem_destroy(&cheio);
}

// Função produtora de números
void* produtor(){
    int numero_gerado;

    while(numeros_produzidos<MAX_MEMORIA_CONSUMIDA) {

        // Gera número aleatório
        numero_gerado = rand()%100000;

        // Espera existir um espaço livre no buffer
        sem_wait(&vazio);
        // Espera pelo acesso ao buffer
        sem_wait(&mutex);

        // Insere número gerado no buffer
        produzir_numero_buffer(numero_gerado);

        // Libera acesso ao buffer
        sem_post(&mutex);
        // Sinaliza que um espaço do buffer foi ocupado
        sem_post(&cheio);
    }
}

// Função consumidora de números
void* consumidor(){
    int numero_para_checar;

    while(numeros_consumidos<MAX_MEMORIA_CONSUMIDA) {

        // Espera existir um espaço ocupado no buffer
        sem_wait(&cheio);
        // Espera pelo acesso ao buffer
        sem_wait(&mutex);

        // Consome um número do buffer
        numero_para_checar = consumir_numero_buffer();

        // Libera acesso ao buffer
        sem_post(&mutex);
        // Sinaliza que um espaço do buffer foi desocupado
        sem_post(&vazio);

        // Se o numero recebido não foi um código de erro
        if(numero_para_checar != -1){

            // Checa se o número é primo
            int bool_primo = checar_numero_primo(numero_para_checar);

            // Informa o usuário
            if(bool_primo){
                printf("O número %ld é primo\n", numero_para_checar);
            } else {
                printf("O número %ld não é primo\n", numero_para_checar);
            }
        }
    }
}

// Escreve o numero gerado no buffer
void produzir_numero_buffer(int numero_gerado){

    // Se o limite de números produzidos não tiver sido atingido
    if(numeros_produzidos < MAX_MEMORIA_CONSUMIDA){   

        // Procura um espaço livre no buffer
        int espaco_livre = procurar_espaco_livre();

        // Informa o número inserido
        printf(" inserindo numero %d\n", numero_gerado);

        // Escreve o número no buffer
        buffer_compartilhado[espaco_livre] = numero_gerado;

        // Incrementa o contador de números já produzidos
        numeros_produzidos++;
    }
}

// Lê um número do buffer
int consumir_numero_buffer(){

    // Código de erro
    int temp = -1;

    // Se o limite de números consumidos não tiver sido atingido
    if(numeros_consumidos < MAX_MEMORIA_CONSUMIDA){

        // Procura um espaço ocupado no buffer
        int espaco_cheio = procurar_espaco_cheio();

        // Lê o número do buffer
        temp = buffer_compartilhado[espaco_cheio];

        // Desocupa o espaço
        buffer_compartilhado[espaco_cheio] = 0;

        // Incrementa o contador de números já consumidos
        numeros_consumidos++;
    
    } // Se o limite for atingido, retorna código de erro

    return temp;
}

// Retorna o primeiro espaço com 0 encontrado
unsigned int procurar_espaco_livre(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(buffer_compartilhado[i] == 0){
            return i;
        }
    }
}

// Retorna o primeiro espaço diferente de 0 encontrado
unsigned int procurar_espaco_cheio(){
    unsigned int i;
    for(i = 0; i < BUFFER_SIZE; i++){
        if(buffer_compartilhado[i] != 0){
            return i;
        }
    }
}

// Retorna 1 sse o número for primo
// Retorna 0 senão
int checar_numero_primo(unsigned int numero){
    unsigned int i;

    if(numero == 0) return 0;

    for(i=2;i<numero;i++){
        if(numero%i == 0) return 0;
    }

    return 1;
}

// Inicia as threads e espera pelo fim da execução
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