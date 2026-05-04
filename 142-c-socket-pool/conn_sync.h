#ifndef CONN_SYNC_H
#define CONN_SYNC_H

#include <pthread.h>
#include <stdbool.h>

typedef struct SyncLock {
    pthread_mutex_t mutex;
} SyncLock;

typedef struct SyncCond {
    pthread_cond_t cond;
    pthread_mutex_t *mutex;
} SyncCond;

void sync_lock_init(SyncLock *lock);
void sync_lock_destroy(SyncLock *lock);
void sync_lock_acquire(SyncLock *lock);
void sync_lock_release(SyncLock *lock);

void sync_cond_init(SyncCond *cond, SyncLock *lock);
void sync_cond_destroy(SyncCond *cond);
void sync_cond_signal(SyncCond *cond);
void sync_cond_broadcast(SyncCond *cond);
int sync_cond_wait_timeout(SyncCond *cond, int timeout_ms);
void sync_cond_wait(SyncCond *cond);

#endif
