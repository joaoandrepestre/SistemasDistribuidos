#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <wait.h>
#include <time.h>

int numeroPrimo(int num){

    int i;
    for(i=2;i<num;i++){
        if(num%i == 0) return 0;
    }

    return 1;
}

int main(int argc, char **argv){

    int qt, fork_ret;

    // Inicia gerador de números aleatórios
    time_t seed;
    srand((unsigned)time(&seed));

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

    // Verifica erro de fork
    if(fork_ret < 0){
        printf("Falha na bifurcação do processo.\n");
        return -1;
    }

    // Processo pai(agindo como produtor)
    if(fork_ret > 0){

        int randNum, i;
        char msgSend[20];

        // Produtor fecha ponta de leitura do pipe
        close(pip[0]);

        // Produz qt números aleatórios
        randNum = 1;
        for(i=0;i<qt;i++){

            // Gera número aleatório crescente
            randNum += rand()%100;
            printf("Número %d gerado.\n", randNum);

            // Define mensagem a ser enviada
            sprintf(msgSend, "%d", randNum);

            // Escreve mensagem na ponta de escrita do pipe
            write(pip[1], msgSend, 20);
        }

        // Escreve mensagem final (0)
        sprintf(msgSend, "%d", 0);
        write(pip[1], msgSend, 20);

        // Fecha ponta de escrita do pipe
        close(pip[1]);

        // Termina processo
        printf("Zero enviado, terminando processo.\n");
        exit(0);
    } 
    // Processo filho (agindo como consumidor)
    else{

        int num;
        char msgRec[20];

        // Consumidor fecha ponta de escrita do pipe
        close(pip[1]);

        // Lê da ponta de leitura do pipe
        read(pip[0], msgRec, 20);
        num = atoi(msgRec);

        // Equanto não receber um 0
        while(num != 0){

            // Informa o número recebido
            printf("Número %d recebido.\n", num);

            // Checa se o número é primo
            if(numeroPrimo(num)) printf("O número %d é primo.\n", num);
            else printf("O número %d não é primo.\n", num);

            // Lê o próximo número da ponta de leitura do pipe
            read(pip[0], msgRec, 20);
            num = atoi(msgRec);
        }

        // Fecha ponta de leitura do pipe
        close(pip[0]);

        // Termina o processo
        printf("Zero recebido, terminando processo.\n");
        exit(0);
    }

    return 0;
}