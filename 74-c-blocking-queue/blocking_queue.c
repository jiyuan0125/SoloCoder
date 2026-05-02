#include "blocking_queue.h"
#include <stdlib.h>

struct blocking_queue {
    void** buffer;
    int capacity;
    int size;
    int front;
    int rear;
    int closed;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
};

blocking_queue_t* bq_create(int capacity) {
    if (capacity <= 0) {
        return NULL;
    }
    blocking_queue_t* q = (blocking_queue_t*)malloc(sizeof(blocking_queue_t));
    if (q == NULL) {
        return NULL;
    }
    q->buffer = (void**)malloc(sizeof(void*) * capacity);
    if (q->buffer == NULL) {
        free(q);
        return NULL;
    }
    q->capacity = capacity;
    q->size = 0;
    q->front = 0;
    q->rear = 0;
    q->closed = 0;
    pthread_mutex_init(&q->mutex, NULL);
    pthread_cond_init(&q->not_empty, NULL);
    pthread_cond_init(&q->not_full, NULL);
    return q;
}

void bq_destroy(blocking_queue_t* queue) {
    if (queue == NULL) {
        return;
    }
    pthread_mutex_lock(&queue->mutex);
    pthread_cond_destroy(&queue->not_empty);
    pthread_cond_destroy(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
    pthread_mutex_destroy(&queue->mutex);
    free(queue->buffer);
    free(queue);
}

int bq_put(blocking_queue_t* queue, void* item) {
    if (queue == NULL) {
        return -1;
    }
    pthread_mutex_lock(&queue->mutex);
    while (queue->size == queue->capacity && !queue->closed) {
        pthread_cond_wait(&queue->not_full, &queue->mutex);
    }
    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return -1;
    }
    queue->buffer[queue->rear] = item;
    queue->rear = (queue->rear + 1) % queue->capacity;
    queue->size++;
    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);
    return 0;
}

int bq_take(blocking_queue_t* queue, void** out_item) {
    if (queue == NULL || out_item == NULL) {
        return -1;
    }
    pthread_mutex_lock(&queue->mutex);
    while (queue->size == 0 && !queue->closed) {
        pthread_cond_wait(&queue->not_empty, &queue->mutex);
    }
    if (queue->size == 0 && queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return -1;
    }
    *out_item = queue->buffer[queue->front];
    queue->front = (queue->front + 1) % queue->capacity;
    queue->size--;
    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
    return 0;
}

void bq_close(blocking_queue_t* queue) {
    if (queue == NULL) {
        return;
    }
    pthread_mutex_lock(&queue->mutex);
    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return;
    }
    queue->closed = 1;
    pthread_cond_broadcast(&queue->not_empty);
    pthread_cond_broadcast(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
}

int bq_size(blocking_queue_t* queue) {
    if (queue == NULL) {
        return 0;
    }
    pthread_mutex_lock(&queue->mutex);
    int size = queue->size;
    pthread_mutex_unlock(&queue->mutex);
    return size;
}

int bq_batch_put(blocking_queue_t* queue, void** items, int count) {
    if (queue == NULL || items == NULL || count <= 0) {
        return 0;
    }
    pthread_mutex_lock(&queue->mutex);
    if (queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return -1;
    }
    if ((queue->capacity - queue->size) < count) {
        pthread_mutex_unlock(&queue->mutex);
        return 0;
    }
    for (int i = 0; i < count; i++) {
        queue->buffer[queue->rear] = items[i];
        queue->rear = (queue->rear + 1) % queue->capacity;
        queue->size++;
    }
    pthread_cond_signal(&queue->not_empty);
    pthread_mutex_unlock(&queue->mutex);
    return count;
}

int bq_batch_take(blocking_queue_t* queue, void** items, int max_count) {
    if (queue == NULL || items == NULL || max_count <= 0) {
        return 0;
    }
    pthread_mutex_lock(&queue->mutex);
    if (queue->size == 0 && queue->closed) {
        pthread_mutex_unlock(&queue->mutex);
        return -1;
    }
    if (queue->size == 0) {
        pthread_mutex_unlock(&queue->mutex);
        return 0;
    }
    int take_count = (queue->size < max_count) ? queue->size : max_count;
    for (int i = 0; i < take_count; i++) {
        items[i] = queue->buffer[queue->front];
        queue->front = (queue->front + 1) % queue->capacity;
        queue->size--;
    }
    pthread_cond_signal(&queue->not_full);
    pthread_mutex_unlock(&queue->mutex);
    return take_count;
}
