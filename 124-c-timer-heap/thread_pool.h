#ifndef THREAD_POOL_H
#define THREAD_POOL_H

#include <pthread.h>
#include "min_heap.h"

#define DEFAULT_THREAD_COUNT 4
#define DEFAULT_QUEUE_SIZE 64

typedef struct pool_task {
    timer_task_t *timer_task;
    struct pool_task *next;
} pool_task_t;

typedef struct thread_pool {
    pool_task_t *queue_head;
    pool_task_t *queue_tail;
    size_t queue_size;
    size_t max_queue_size;
    
    pthread_t *threads;
    size_t thread_count;
    
    pthread_mutex_t queue_mutex;
    pthread_cond_t queue_not_empty;
    pthread_cond_t queue_not_full;
    
    int shutdown;
} thread_pool_t;

thread_pool_t* thread_pool_create(size_t thread_count, size_t max_queue_size);
void thread_pool_destroy(thread_pool_t *pool);
int thread_pool_submit(thread_pool_t *pool, timer_task_t *task);
size_t thread_pool_pending_count(thread_pool_t *pool);

#endif
