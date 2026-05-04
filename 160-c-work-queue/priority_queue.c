#include "priority_queue.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int priority_queue_init(PriorityQueue *pq, int total_capacity) {
    if (pq == NULL || total_capacity <= 0) {
        return -1;
    }
    
    pq->total_capacity = total_capacity;
    pq->shutdown = 0;
    
    int per_priority_cap = (total_capacity + PRIORITY_LEVELS - 1) / PRIORITY_LEVELS;
    
    for (int i = 0; i < PRIORITY_LEVELS; i++) {
        if (task_queue_init(&pq->queues[i], per_priority_cap) != 0) {
            for (int j = 0; j < i; j++) {
                task_queue_destroy(&pq->queues[j]);
            }
            return -1;
        }
    }
    
    if (pthread_mutex_init(&pq->mutex, NULL) != 0) {
        for (int i = 0; i < PRIORITY_LEVELS; i++) {
            task_queue_destroy(&pq->queues[i]);
        }
        return -1;
    }
    
    if (pthread_cond_init(&pq->not_empty, NULL) != 0) {
        pthread_mutex_destroy(&pq->mutex);
        for (int i = 0; i < PRIORITY_LEVELS; i++) {
            task_queue_destroy(&pq->queues[i]);
        }
        return -1;
    }
    
    return 0;
}

void priority_queue_destroy(PriorityQueue *pq) {
    if (pq == NULL) {
        return;
    }
    
    priority_queue_clear(pq);
    
    for (int i = 0; i < PRIORITY_LEVELS; i++) {
        task_queue_destroy(&pq->queues[i]);
    }
    
    pthread_mutex_destroy(&pq->mutex);
    pthread_cond_destroy(&pq->not_empty);
}

static int priority_queue_total_size_locked(PriorityQueue *pq) {
    int total = 0;
    for (int i = 0; i < PRIORITY_LEVELS; i++) {
        total += task_queue_size(&pq->queues[i]);
    }
    return total;
}

int priority_queue_enqueue(PriorityQueue *pq, Task *task) {
    if (pq == NULL || task == NULL) {
        return -1;
    }
    
    pthread_mutex_lock(&pq->mutex);
    
    if (pq->shutdown) {
        pthread_mutex_unlock(&pq->mutex);
        return -1;
    }
    
    int total_size = priority_queue_total_size_locked(pq);
    if (total_size >= pq->total_capacity) {
        pthread_mutex_unlock(&pq->mutex);
        return -2;
    }
    
    TaskPriority prio = task->priority;
    if (prio < TASK_PRIORITY_LOW || prio > TASK_PRIORITY_HIGH) {
        prio = TASK_PRIORITY_LOW;
    }
    
    if (task_queue_enqueue(&pq->queues[prio], task) != 0) {
        pthread_mutex_unlock(&pq->mutex);
        return -1;
    }
    
    pthread_cond_signal(&pq->not_empty);
    pthread_mutex_unlock(&pq->mutex);
    
    return 0;
}

Task* priority_queue_dequeue(PriorityQueue *pq) {
    if (pq == NULL) {
        return NULL;
    }
    
    pthread_mutex_lock(&pq->mutex);
    
    while (!pq->shutdown) {
        for (int i = PRIORITY_LEVELS - 1; i >= 0; i--) {
            Task *task = task_queue_try_dequeue(&pq->queues[i]);
            if (task != NULL) {
                pthread_mutex_unlock(&pq->mutex);
                return task;
            }
        }
        
        pthread_cond_wait(&pq->not_empty, &pq->mutex);
    }
    
    pthread_mutex_unlock(&pq->mutex);
    return NULL;
}

Task* priority_queue_try_dequeue(PriorityQueue *pq) {
    if (pq == NULL) {
        return NULL;
    }
    
    pthread_mutex_lock(&pq->mutex);
    
    for (int i = PRIORITY_LEVELS - 1; i >= 0; i--) {
        Task *task = task_queue_try_dequeue(&pq->queues[i]);
        if (task != NULL) {
            pthread_mutex_unlock(&pq->mutex);
            return task;
        }
    }
    
    pthread_mutex_unlock(&pq->mutex);
    return NULL;
}

int priority_queue_total_size(PriorityQueue *pq) {
    if (pq == NULL) {
        return -1;
    }
    
    pthread_mutex_lock(&pq->mutex);
    int total = priority_queue_total_size_locked(pq);
    pthread_mutex_unlock(&pq->mutex);
    
    return total;
}

int priority_queue_size_by_priority(PriorityQueue *pq, TaskPriority priority) {
    if (pq == NULL) {
        return -1;
    }
    
    if (priority < TASK_PRIORITY_LOW || priority > TASK_PRIORITY_HIGH) {
        return -1;
    }
    
    return task_queue_size(&pq->queues[priority]);
}

int priority_queue_is_empty(PriorityQueue *pq) {
    if (pq == NULL) {
        return 1;
    }
    
    pthread_mutex_lock(&pq->mutex);
    int empty = (priority_queue_total_size_locked(pq) == 0);
    pthread_mutex_unlock(&pq->mutex);
    
    return empty;
}

int priority_queue_is_full(PriorityQueue *pq) {
    if (pq == NULL) {
        return 1;
    }
    
    pthread_mutex_lock(&pq->mutex);
    int full = (priority_queue_total_size_locked(pq) >= pq->total_capacity);
    pthread_mutex_unlock(&pq->mutex);
    
    return full;
}

void priority_queue_shutdown(PriorityQueue *pq) {
    if (pq == NULL) {
        return;
    }
    
    pthread_mutex_lock(&pq->mutex);
    pq->shutdown = 1;
    pthread_cond_broadcast(&pq->not_empty);
    pthread_mutex_unlock(&pq->mutex);
}

int priority_queue_is_shutdown(PriorityQueue *pq) {
    if (pq == NULL) {
        return 1;
    }
    
    int shutdown;
    pthread_mutex_lock(&pq->mutex);
    shutdown = pq->shutdown;
    pthread_mutex_unlock(&pq->mutex);
    
    return shutdown;
}

void priority_queue_clear(PriorityQueue *pq) {
    if (pq == NULL) {
        return;
    }
    
    for (int i = 0; i < PRIORITY_LEVELS; i++) {
        task_queue_clear(&pq->queues[i]);
    }
}
