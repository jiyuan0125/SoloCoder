#ifndef THREAD_POOL_H
#define THREAD_POOL_H

#include <pthread.h>

typedef struct task {
    void (*function)(void *);
    void *argument;
    struct task *next;
} task_t;

typedef struct {
    pthread_mutex_t lock;
    pthread_cond_t notify;
    pthread_t *threads;
    task_t *head;
    task_t *tail;
    int thread_count;
    int task_count;
    int shutdown;
    int started;
} thread_pool_t;

thread_pool_t *thread_pool_create(int num_threads);
int thread_pool_add(thread_pool_t *pool, void (*function)(void *), void *argument);
int thread_pool_destroy(thread_pool_t *pool);

#endif
