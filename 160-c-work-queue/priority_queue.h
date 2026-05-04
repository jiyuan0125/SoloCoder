#ifndef PRIORITY_QUEUE_H
#define PRIORITY_QUEUE_H

#include "task_queue.h"

#define PRIORITY_LEVELS 3

typedef struct PriorityQueue {
    TaskQueue queues[PRIORITY_LEVELS];
    int total_capacity;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    int shutdown;
} PriorityQueue;

int priority_queue_init(PriorityQueue *pq, int total_capacity);

void priority_queue_destroy(PriorityQueue *pq);

int priority_queue_enqueue(PriorityQueue *pq, Task *task);

Task* priority_queue_dequeue(PriorityQueue *pq);

Task* priority_queue_try_dequeue(PriorityQueue *pq);

int priority_queue_total_size(PriorityQueue *pq);

int priority_queue_size_by_priority(PriorityQueue *pq, TaskPriority priority);

int priority_queue_is_empty(PriorityQueue *pq);

int priority_queue_is_full(PriorityQueue *pq);

void priority_queue_shutdown(PriorityQueue *pq);

int priority_queue_is_shutdown(PriorityQueue *pq);

void priority_queue_clear(PriorityQueue *pq);

#endif
