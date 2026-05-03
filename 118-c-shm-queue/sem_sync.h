#ifndef SEM_SYNC_H
#define SEM_SYNC_H

#include <stdbool.h>
#include <semaphore.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

#define SEM_SYNC_WAIT_INFINITE (-1)
#define SEM_SYNC_WAIT_NOWAIT   (0)

typedef struct sem_sync {
    const char *name;
    sem_t *sem;
    bool owner;
} sem_sync_t;

int sem_sync_create(sem_sync_t *sync, const char *name, unsigned int value);
int sem_sync_open(sem_sync_t *sync, const char *name);
int sem_sync_close(sem_sync_t *sync);
int sem_sync_unlink(const char *name);
bool sem_sync_exists(const char *name);
void sem_sync_init(sem_sync_t *sync);

int sem_sync_wait(sem_sync_t *sync);
int sem_sync_trywait(sem_sync_t *sync);
int sem_sync_timedwait(sem_sync_t *sync, int timeout_ms);
int sem_sync_post(sem_sync_t *sync);
int sem_sync_getvalue(sem_sync_t *sync, int *value);

typedef struct sem_set {
    sem_sync_t mutex;
    sem_sync_t empty;
    sem_sync_t full;
    const char *base_name;
} sem_set_t;

int sem_set_create(sem_set_t *set, const char *base_name, 
                   unsigned int mutex_val, unsigned int empty_val, unsigned int full_val);
int sem_set_open(sem_set_t *set, const char *base_name);
int sem_set_close(sem_set_t *set);
int sem_set_unlink(const char *base_name);
void sem_set_init(sem_set_t *set);
bool sem_set_exists(const char *base_name);

#ifdef __cplusplus
}
#endif

#endif
