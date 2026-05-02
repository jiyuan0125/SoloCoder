#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include "thread_pool.h"

static void *thread_worker(void *arg) {
    thread_pool_t *pool = (thread_pool_t *)arg;
    task_t *task;

    while (1) {
        pthread_mutex_lock(&(pool->lock));

        while ((pool->task_count == 0) && (!pool->shutdown)) {
            pthread_cond_wait(&(pool->notify), &(pool->lock));
        }

        if (pool->shutdown) {
            break;
        }

        task = pool->head;
        pool->head = task->next;
        pool->task_count--;

        pthread_mutex_unlock(&(pool->lock));

        (*(task->function))(task->argument);
        free(task);
    }

    pool->started--;
    pthread_mutex_unlock(&(pool->lock));
    pthread_exit(NULL);
    return NULL;
}

thread_pool_t *thread_pool_create(int num_threads) {
    if (num_threads <= 0) {
        num_threads = 1;
    }

    thread_pool_t *pool = (thread_pool_t *)calloc(1, sizeof(thread_pool_t));
    if (!pool) {
        return NULL;
    }

    pool->thread_count = num_threads;
    pool->threads = (pthread_t *)calloc(num_threads, sizeof(pthread_t));
    if (!pool->threads) {
        free(pool);
        return NULL;
    }

    if (pthread_mutex_init(&(pool->lock), NULL) != 0 ||
        pthread_cond_init(&(pool->notify), NULL) != 0) {
        free(pool->threads);
        free(pool);
        return NULL;
    }

    for (int i = 0; i < num_threads; i++) {
        if (pthread_create(&(pool->threads[i]), NULL, thread_worker, (void *)pool) != 0) {
            thread_pool_destroy(pool);
            return NULL;
        }
        pool->started++;
    }

    return pool;
}

int thread_pool_add(thread_pool_t *pool, void (*function)(void *), void *argument) {
    if (!pool || !function) {
        return -1;
    }

    task_t *task = (task_t *)calloc(1, sizeof(task_t));
    if (!task) {
        return -1;
    }
    task->function = function;
    task->argument = argument;

    pthread_mutex_lock(&(pool->lock));

    if (pool->shutdown) {
        free(task);
        pthread_mutex_unlock(&(pool->lock));
        return -1;
    }

    task->next = NULL;
    if (pool->tail == NULL) {
        pool->head = task;
        pool->tail = task;
    } else {
        pool->tail->next = task;
        pool->tail = task;
    }
    pool->task_count++;

    pthread_cond_signal(&(pool->notify));
    pthread_mutex_unlock(&(pool->lock));

    return 0;
}

int thread_pool_destroy(thread_pool_t *pool) {
    if (!pool) {
        return -1;
    }

    pthread_mutex_lock(&(pool->lock));

    pool->shutdown = 1;

    pthread_cond_broadcast(&(pool->notify));
    pthread_mutex_unlock(&(pool->lock));

    for (int i = 0; i < pool->thread_count; i++) {
        pthread_join(pool->threads[i], NULL);
    }

    task_t *task;
    while (pool->head) {
        task = pool->head;
        pool->head = task->next;
        free(task);
    }

    free(pool->threads);
    pthread_mutex_destroy(&(pool->lock));
    pthread_cond_destroy(&(pool->notify));
    free(pool);

    return 0;
}
