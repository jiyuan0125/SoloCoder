#include "thread_pool.h"
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

static void* worker_thread(void *arg) {
    thread_pool_t *pool = (thread_pool_t *)arg;
    
    while (1) {
        pthread_mutex_lock(&pool->queue_mutex);
        
        while (pool->queue_size == 0 && !pool->shutdown) {
            pthread_cond_wait(&pool->queue_not_empty, &pool->queue_mutex);
        }
        
        if (pool->shutdown) {
            pthread_mutex_unlock(&pool->queue_mutex);
            break;
        }
        
        pool_task_t *task_node = pool->queue_head;
        pool->queue_head = task_node->next;
        pool->queue_size--;
        
        if (pool->queue_head == NULL) {
            pool->queue_tail = NULL;
        }
        
        pthread_cond_signal(&pool->queue_not_full);
        pthread_mutex_unlock(&pool->queue_mutex);
        
        timer_task_t *task = task_node->timer_task;
        free(task_node);
        
        if (task && task->callback) {
            if (!task->cancelled) {
                task->executed_count++;
                task->callback(task->arg);
            }
        }
    }
    
    return NULL;
}

thread_pool_t* thread_pool_create(size_t thread_count, size_t max_queue_size) {
    if (thread_count == 0) {
        thread_count = DEFAULT_THREAD_COUNT;
    }
    if (max_queue_size == 0) {
        max_queue_size = DEFAULT_QUEUE_SIZE;
    }
    
    thread_pool_t *pool = (thread_pool_t *)malloc(sizeof(thread_pool_t));
    if (pool == NULL) {
        return NULL;
    }
    
    pool->queue_head = NULL;
    pool->queue_tail = NULL;
    pool->queue_size = 0;
    pool->max_queue_size = max_queue_size;
    pool->shutdown = 0;
    
    pool->thread_count = thread_count;
    pool->threads = (pthread_t *)malloc(thread_count * sizeof(pthread_t));
    if (pool->threads == NULL) {
        free(pool);
        return NULL;
    }
    
    if (pthread_mutex_init(&pool->queue_mutex, NULL) != 0) {
        free(pool->threads);
        free(pool);
        return NULL;
    }
    
    if (pthread_cond_init(&pool->queue_not_empty, NULL) != 0) {
        pthread_mutex_destroy(&pool->queue_mutex);
        free(pool->threads);
        free(pool);
        return NULL;
    }
    
    if (pthread_cond_init(&pool->queue_not_full, NULL) != 0) {
        pthread_cond_destroy(&pool->queue_not_empty);
        pthread_mutex_destroy(&pool->queue_mutex);
        free(pool->threads);
        free(pool);
        return NULL;
    }
    
    for (size_t i = 0; i < thread_count; i++) {
        if (pthread_create(&pool->threads[i], NULL, worker_thread, pool) != 0) {
            for (size_t j = 0; j < i; j++) {
                pthread_join(pool->threads[j], NULL);
            }
            pthread_cond_destroy(&pool->queue_not_full);
            pthread_cond_destroy(&pool->queue_not_empty);
            pthread_mutex_destroy(&pool->queue_mutex);
            free(pool->threads);
            free(pool);
            return NULL;
        }
    }
    
    return pool;
}

void thread_pool_destroy(thread_pool_t *pool) {
    if (pool == NULL) {
        return;
    }
    
    pthread_mutex_lock(&pool->queue_mutex);
    pool->shutdown = 1;
    pthread_cond_broadcast(&pool->queue_not_empty);
    pthread_mutex_unlock(&pool->queue_mutex);
    
    for (size_t i = 0; i < pool->thread_count; i++) {
        pthread_join(pool->threads[i], NULL);
    }
    
    while (pool->queue_head != NULL) {
        pool_task_t *temp = pool->queue_head;
        pool->queue_head = temp->next;
        free(temp);
    }
    
    free(pool->threads);
    pthread_mutex_destroy(&pool->queue_mutex);
    pthread_cond_destroy(&pool->queue_not_empty);
    pthread_cond_destroy(&pool->queue_not_full);
    free(pool);
}

int thread_pool_submit(thread_pool_t *pool, timer_task_t *task) {
    if (pool == NULL || task == NULL) {
        return -1;
    }
    
    pool_task_t *task_node = (pool_task_t *)malloc(sizeof(pool_task_t));
    if (task_node == NULL) {
        return -1;
    }
    task_node->timer_task = task;
    task_node->next = NULL;
    
    pthread_mutex_lock(&pool->queue_mutex);
    
    while (pool->queue_size >= pool->max_queue_size && !pool->shutdown) {
        pthread_cond_wait(&pool->queue_not_full, &pool->queue_mutex);
    }
    
    if (pool->shutdown) {
        pthread_mutex_unlock(&pool->queue_mutex);
        free(task_node);
        return -1;
    }
    
    if (pool->queue_tail == NULL) {
        pool->queue_head = task_node;
        pool->queue_tail = task_node;
    } else {
        pool->queue_tail->next = task_node;
        pool->queue_tail = task_node;
    }
    pool->queue_size++;
    
    pthread_cond_signal(&pool->queue_not_empty);
    pthread_mutex_unlock(&pool->queue_mutex);
    
    return 0;
}

size_t thread_pool_pending_count(thread_pool_t *pool) {
    if (pool == NULL) {
        return 0;
    }
    pthread_mutex_lock(&pool->queue_mutex);
    size_t count = pool->queue_size;
    pthread_mutex_unlock(&pool->queue_mutex);
    return count;
}
