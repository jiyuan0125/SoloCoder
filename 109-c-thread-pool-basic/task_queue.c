#include "task_queue.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>

tp_status_t tp_task_queue_init(tp_task_queue_t *queue, size_t capacity)
{
    if (!queue || capacity == 0) {
        return TP_INVALID_ARG;
    }

    queue->buffer = (tp_task_t *)calloc(capacity, sizeof(tp_task_t));
    if (!queue->buffer) {
        return TP_ERROR;
    }

    queue->capacity = capacity;
    queue->head = 0;
    queue->tail = 0;
    queue->count = 0;
    queue->closed = 0;

    if (pthread_mutex_init(&queue->mutex, NULL) != 0) {
        free(queue->buffer);
        return TP_ERROR;
    }

    if (pthread_cond_init(&queue->not_empty, NULL) != 0) {
        pthread_mutex_destroy(&queue->mutex);
        free(queue->buffer);
        return TP_ERROR;
    }

    if (pthread_cond_init(&queue->not_full, NULL) != 0) {
        pthread_cond_destroy(&queue->not_empty);
        pthread_mutex_destroy(&queue->mutex);
        free(queue->buffer);
        return TP_ERROR;
    }

    return TP_OK;
}

void tp_task_queue_destroy(tp_task_queue_t *queue)
{
    if (!queue) return;

    pthread_mutex_lock(&queue->mutex);
    if (queue->buffer) {
        free(queue->buffer);
        queue->buffer = NULL;
    }
    pthread_mutex_unlock(&queue->mutex);

    pthread_cond_destroy(&queue->not_full);
    pthread_cond_destroy(&queue->not_empty);
    pthread_mutex_destroy(&queue->mutex);
}

tp_status_t tp_task_queue_push(tp_task_queue_t *queue, const tp_task_t *task, int block)
{
    if (!queue || !task) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&queue->mutex);

    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_CLOSED;
    }

    while (queue->count >= queue->capacity && !queue->closed) {
        if (!block) {
            pthread_mutex_unlock(&queue->mutex);
            return TP_FULL;
        }
        pthread_cond_wait(&queue->not_full, &queue->mutex);
    }

    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_CLOSED;
    }

    memcpy(&queue->buffer[queue->tail], task, sizeof(tp_task_t));
    queue->tail = (queue->tail + 1) % queue->capacity;
    queue->count++;

    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);

    return TP_OK;
}

tp_status_t tp_task_queue_pop(tp_task_queue_t *queue, tp_task_t *task)
{
    if (!queue || !task) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&queue->mutex);

    while (queue->count == 0 && !queue->closed) {
        pthread_cond_wait(&queue->not_empty, &queue->mutex);
    }

    if (queue->count == 0 && queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_CLOSED;
    }

    memcpy(task, &queue->buffer[queue->head], sizeof(tp_task_t));
    queue->head = (queue->head + 1) % queue->capacity;
    queue->count--;

    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);

    return TP_OK;
}

tp_status_t tp_task_queue_try_pop(tp_task_queue_t *queue, tp_task_t *task)
{
    if (!queue || !task) {
        return TP_INVALID_ARG;
    }

    pthread_mutex_lock(&queue->mutex);

    if (queue->count == 0) {
        pthread_mutex_unlock(&queue->mutex);
        return TP_FULL;
    }

    memcpy(task, &queue->buffer[queue->head], sizeof(tp_task_t));
    queue->head = (queue->head + 1) % queue->capacity;
    queue->count--;

    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);

    return TP_OK;
}

void tp_task_queue_close(tp_task_queue_t *queue)
{
    if (!queue) return;

    pthread_mutex_lock(&queue->mutex);
    queue->closed = 1;
    pthread_cond_broadcast(&queue->not_empty);
    pthread_cond_broadcast(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
}

size_t tp_task_queue_size(const tp_task_queue_t *queue)
{
    if (!queue) return 0;

    pthread_mutex_lock((pthread_mutex_t *)&queue->mutex);
    size_t size = queue->count;
    pthread_mutex_unlock((pthread_mutex_t *)&queue->mutex);

    return size;
}

int tp_task_queue_is_full(const tp_task_queue_t *queue)
{
    if (!queue) return 1;

    pthread_mutex_lock((pthread_mutex_t *)&queue->mutex);
    int full = (queue->count >= queue->capacity);
    pthread_mutex_unlock((pthread_mutex_t *)&queue->mutex);

    return full;
}

int tp_task_queue_is_empty(const tp_task_queue_t *queue)
{
    if (!queue) return 1;

    pthread_mutex_lock((pthread_mutex_t *)&queue->mutex);
    int empty = (queue->count == 0);
    pthread_mutex_unlock((pthread_mutex_t *)&queue->mutex);

    return empty;
}
