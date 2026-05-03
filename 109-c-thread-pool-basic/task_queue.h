#ifndef TASK_QUEUE_H
#define TASK_QUEUE_H

#include "common.h"
#include <pthread.h>

typedef struct {
    tp_task_t *buffer;
    size_t capacity;
    size_t head;
    size_t tail;
    size_t count;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
    int closed;
} tp_task_queue_t;

tp_status_t tp_task_queue_init(tp_task_queue_t *queue, size_t capacity);
void tp_task_queue_destroy(tp_task_queue_t *queue);
tp_status_t tp_task_queue_push(tp_task_queue_t *queue, const tp_task_t *task, int block);
tp_status_t tp_task_queue_pop(tp_task_queue_t *queue, tp_task_t *task);
tp_status_t tp_task_queue_try_pop(tp_task_queue_t *queue, tp_task_t *task);
void tp_task_queue_close(tp_task_queue_t *queue);
size_t tp_task_queue_size(const tp_task_queue_t *queue);
int tp_task_queue_is_full(const tp_task_queue_t *queue);
int tp_task_queue_is_empty(const tp_task_queue_t *queue);

#endif
