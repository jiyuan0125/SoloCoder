#include "conn_sync.h"
#include <errno.h>
#include <sys/time.h>
#include <stddef.h>

void sync_lock_init(SyncLock *lock)
{
    pthread_mutex_init(&lock->mutex, NULL);
}

void sync_lock_destroy(SyncLock *lock)
{
    pthread_mutex_destroy(&lock->mutex);
}

void sync_lock_acquire(SyncLock *lock)
{
    pthread_mutex_lock(&lock->mutex);
}

void sync_lock_release(SyncLock *lock)
{
    pthread_mutex_unlock(&lock->mutex);
}

void sync_cond_init(SyncCond *cond, SyncLock *lock)
{
    pthread_cond_init(&cond->cond, NULL);
    cond->mutex = &lock->mutex;
}

void sync_cond_destroy(SyncCond *cond)
{
    pthread_cond_destroy(&cond->cond);
}

void sync_cond_signal(SyncCond *cond)
{
    pthread_cond_signal(&cond->cond);
}

void sync_cond_broadcast(SyncCond *cond)
{
    pthread_cond_broadcast(&cond->cond);
}

int sync_cond_wait_timeout(SyncCond *cond, int timeout_ms)
{
    struct timespec ts;
    struct timeval tv;
    
    gettimeofday(&tv, NULL);
    
    ts.tv_sec = tv.tv_sec + timeout_ms / 1000;
    ts.tv_nsec = tv.tv_usec * 1000 + (timeout_ms % 1000) * 1000000;
    
    if (ts.tv_nsec >= 1000000000) {
        ts.tv_sec += ts.tv_nsec / 1000000000;
        ts.tv_nsec = ts.tv_nsec % 1000000000;
    }
    
    return pthread_cond_timedwait(&cond->cond, cond->mutex, &ts);
}

void sync_cond_wait(SyncCond *cond)
{
    pthread_cond_wait(&cond->cond, cond->mutex);
}
