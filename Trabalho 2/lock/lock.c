#include "lock.h"
#include <stdlib.h>

// Construtor para lock
Lock* criaLock(){
    Lock* lock = (Lock*) malloc(sizeof(Lock));
    lock->held = 0;
    return lock;
}

// Implementa função de acquire lock
void acquire(Lock *lock){
    while(__sync_lock_test_and_set(&(lock->held), 1));
}

// Implementa função de release lock
void release(Lock *lock){
    lock->held = 0;
}