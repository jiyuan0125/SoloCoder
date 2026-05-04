#ifndef POOL_H
#define POOL_H

#include "monster.h"
#include "stats.h"
#include <pthread.h>

#define DEFAULT_POOL_SIZE 1000
#define DEFAULT_LEAK_TIMEOUT 300

typedef struct MonsterPool {
    Monster *monsters;
    unsigned int *free_indices;
    unsigned int capacity;
    unsigned int free_count;
    unsigned int used_count;
    unsigned int leak_timeout_seconds;
    pthread_mutex_t mutex;
    PoolStats stats;
    int initialized;
} MonsterPool;

int pool_create(MonsterPool *pool, unsigned int capacity, unsigned int leak_timeout_seconds);
void pool_destroy(MonsterPool *pool);
Monster* pool_acquire(MonsterPool *pool);
int pool_release(MonsterPool *pool, Monster *monster);
unsigned int pool_get_free_count(MonsterPool *pool);
unsigned int pool_get_used_count(MonsterPool *pool);
unsigned int pool_get_capacity(MonsterPool *pool);
unsigned int pool_recover_leaks(MonsterPool *pool);
void pool_get_statistics(MonsterPool *pool, unsigned long long *total_alloc, 
                         unsigned long long *total_dealloc, unsigned long long *leak_recoveries);

#endif
