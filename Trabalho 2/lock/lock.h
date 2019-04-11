#ifndef LOCK_H
#define LOCK_H

// Implementaestrutura de lock
typedef struct lock{
    volatile int held;
} Lock;

// Construtor para lock
Lock* criaLock();

// Implementa função de acquire lock
void acquire(Lock *lock);

// Implementa função de release lock
void release(Lock *lock);

#endif //LOCK_H