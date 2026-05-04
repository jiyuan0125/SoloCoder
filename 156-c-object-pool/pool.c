#include "pool.h"
#include "reset.h"
#include <stdlib.h>
#include <string.h>

int pool_create(MonsterPool *pool, unsigned int capacity, unsigned int leak_timeout_seconds) {
    unsigned int i;
    
    if (pool == NULL || capacity == 0) {
        return -1;
    }
    
    memset(pool, 0, sizeof(MonsterPool));
    
    pool->monsters = (Monster*)malloc(sizeof(Monster) * capacity);
    if (pool->monsters == NULL) {
        return -1;
    }
    
    pool->free_indices = (unsigned int*)malloc(sizeof(unsigned int) * capacity);
    if (pool->free_indices == NULL) {
        free(pool->monsters);
        pool->monsters = NULL;
        return -1;
    }
    
    for (i = 0; i < capacity; i++) {
        monster_init(&pool->monsters[i], i);
        pool->free_indices[i] = i;
    }
    
    pool->capacity = capacity;
    pool->free_count = capacity;
    pool->used_count = 0;
    pool->leak_timeout_seconds = leak_timeout_seconds;
    
    pthread_mutex_init(&pool->mutex, NULL);
    stats_init(&pool->stats);
    
    pool->initialized = 1;
    
    return 0;
}

void pool_destroy(MonsterPool *pool) {
    if (pool == NULL || !pool->initialized) {
        return;
    }
    
    pthread_mutex_lock(&pool->mutex);
    
    if (pool->monsters != NULL) {
        free(pool->monsters);
        pool->monsters = NULL;
    }
    
    if (pool->free_indices != NULL) {
        free(pool->free_indices);
        pool->free_indices = NULL;
    }
    
    pool->capacity = 0;
    pool->free_count = 0;
    pool->used_count = 0;
    
    pthread_mutex_unlock(&pool->mutex);
    
    pthread_mutex_destroy(&pool->mutex);
    stats_destroy(&pool->stats);
    
    pool->initialized = 0;
}

Monster* pool_acquire(MonsterPool *pool) {
    Monster *result = NULL;
    unsigned int index;
    unsigned long long current_time;
    
    if (pool == NULL || !pool->initialized) {
        return NULL;
    }
    
    pthread_mutex_lock(&pool->mutex);
    
    if (pool->free_count > 0) {
        pool->free_count--;
        index = pool->free_indices[pool->free_count];
        result = &pool->monsters[index];
        pool->used_count++;
        
        current_time = get_current_timestamp_seconds();
        result->borrow_timestamp = current_time;
        result->state = MONSTER_STATE_ALIVE;
        
        stats_record_allocation(&pool->stats);
    }
    
    pthread_mutex_unlock(&pool->mutex);
    
    return result;
}

int pool_release(MonsterPool *pool, Monster *monster) {
    unsigned int index;
    
    if (pool == NULL || monster == NULL || !pool->initialized) {
        return -1;
    }
    
    pthread_mutex_lock(&pool->mutex);
    
    index = monster->pool_index;
    
    if (index >= pool->capacity) {
        pthread_mutex_unlock(&pool->mutex);
        return -1;
    }
    
    if (monster != &pool->monsters[index]) {
        pthread_mutex_unlock(&pool->mutex);
        return -1;
    }
    
    if (monster->state == MONSTER_STATE_IDLE) {
        pthread_mutex_unlock(&pool->mutex);
        return -1;
    }
    
    monster_reset(monster);
    monster->pool_index = index;
    
    pool->free_indices[pool->free_count] = index;
    pool->free_count++;
    pool->used_count--;
    
    stats_record_deallocation(&pool->stats);
    
    pthread_mutex_unlock(&pool->mutex);
    
    return 0;
}

unsigned int pool_get_free_count(MonsterPool *pool) {
    unsigned int result;
    
    if (pool == NULL || !pool->initialized) {
        return 0;
    }
    
    pthread_mutex_lock(&pool->mutex);
    result = pool->free_count;
    pthread_mutex_unlock(&pool->mutex);
    
    return result;
}

unsigned int pool_get_used_count(MonsterPool *pool) {
    unsigned int result;
    
    if (pool == NULL || !pool->initialized) {
        return 0;
    }
    
    pthread_mutex_lock(&pool->mutex);
    result = pool->used_count;
    pthread_mutex_unlock(&pool->mutex);
    
    return result;
}

unsigned int pool_get_capacity(MonsterPool *pool) {
    if (pool == NULL || !pool->initialized) {
        return 0;
    }
    return pool->capacity;
}

unsigned int pool_recover_leaks(MonsterPool *pool) {
    unsigned int i;
    unsigned int recovered_count = 0;
    unsigned long long current_time;
    
    if (pool == NULL || !pool->initialized) {
        return 0;
    }
    
    current_time = get_current_timestamp_seconds();
    
    pthread_mutex_lock(&pool->mutex);
    
    for (i = 0; i < pool->capacity; i++) {
        Monster *monster = &pool->monsters[i];
        
        if (monster->state != MONSTER_STATE_IDLE) {
            if (monster_is_leaked(monster, current_time, pool->leak_timeout_seconds)) {
                monster_reset(monster);
                monster->pool_index = i;
                
                pool->free_indices[pool->free_count] = i;
                pool->free_count++;
                pool->used_count--;
                
                stats_record_leak_recovery(&pool->stats);
                recovered_count++;
            }
        }
    }
    
    pthread_mutex_unlock(&pool->mutex);
    
    return recovered_count;
}

void pool_get_statistics(MonsterPool *pool, unsigned long long *total_alloc, 
                         unsigned long long *total_dealloc, unsigned long long *leak_recoveries) {
    if (pool == NULL || !pool->initialized) {
        if (total_alloc != NULL) *total_alloc = 0;
        if (total_dealloc != NULL) *total_dealloc = 0;
        if (leak_recoveries != NULL) *leak_recoveries = 0;
        return;
    }
    
    if (total_alloc != NULL) {
        *total_alloc = stats_get_total_allocations(&pool->stats);
    }
    if (total_dealloc != NULL) {
        *total_dealloc = stats_get_total_deallocations(&pool->stats);
    }
    if (leak_recoveries != NULL) {
        *leak_recoveries = stats_get_leak_recoveries(&pool->stats);
    }
}
