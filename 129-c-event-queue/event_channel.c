#include "event_channel.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>

static event_id_t g_next_event_id = 1;
static pthread_mutex_t g_id_mutex = PTHREAD_MUTEX_INITIALIZER;

event_id_t generate_event_id(void) {
    event_id_t id;
    pthread_mutex_lock(&g_id_mutex);
    id = g_next_event_id++;
    pthread_mutex_unlock(&g_id_mutex);
    return id;
}

static event_queue_t* event_queue_create(size_t capacity) {
    event_queue_t *q = (event_queue_t*)calloc(1, sizeof(event_queue_t));
    if (!q) return NULL;
    
    q->heap = (event_t**)calloc(capacity, sizeof(event_t*));
    if (!q->heap) {
        free(q);
        return NULL;
    }
    
    q->capacity = capacity;
    q->size = 0;
    q->sequence_counter = 0;
    return q;
}

static void event_queue_destroy(event_queue_t *q) {
    if (!q) return;
    for (size_t i = 0; i < q->size; i++) {
        event_free(q->heap[i]);
    }
    free(q->heap);
    free(q);
}

static int event_compare(const event_t *a, const event_t *b) {
    if (a->priority != b->priority) {
        return a->priority - b->priority;
    }
    return a->sequence - b->sequence;
}

static void event_queue_heapify_up(event_queue_t *q, size_t idx) {
    while (idx > 0) {
        size_t parent = (idx - 1) / 2;
        if (event_compare(q->heap[idx], q->heap[parent]) >= 0) {
            break;
        }
        event_t *tmp = q->heap[idx];
        q->heap[idx] = q->heap[parent];
        q->heap[parent] = tmp;
        idx = parent;
    }
}

static void event_queue_heapify_down(event_queue_t *q, size_t idx) {
    for (;;) {
        size_t left = 2 * idx + 1;
        size_t right = 2 * idx + 2;
        size_t smallest = idx;
        
        if (left < q->size && event_compare(q->heap[left], q->heap[smallest]) < 0) {
            smallest = left;
        }
        if (right < q->size && event_compare(q->heap[right], q->heap[smallest]) < 0) {
            smallest = right;
        }
        if (smallest == idx) {
            break;
        }
        
        event_t *tmp = q->heap[idx];
        q->heap[idx] = q->heap[smallest];
        q->heap[smallest] = tmp;
        idx = smallest;
    }
}

static int event_queue_push(event_queue_t *q, event_t *event) {
    if (q->size >= q->capacity) {
        return -1;
    }
    
    q->heap[q->size] = event;
    event->sequence = q->sequence_counter++;
    event_queue_heapify_up(q, q->size);
    q->size++;
    return 0;
}

static event_t* event_queue_pop(event_queue_t *q) {
    if (q->size == 0) {
        return NULL;
    }
    
    event_t *result = q->heap[0];
    q->heap[0] = q->heap[q->size - 1];
    q->size--;
    q->heap[q->size] = NULL;
    
    if (q->size > 0) {
        event_queue_heapify_down(q, 0);
    }
    return result;
}

static event_t* event_queue_peek(event_queue_t *q) {
    if (q->size == 0) {
        return NULL;
    }
    return q->heap[0];
}

static int event_queue_erase(event_queue_t *q, event_id_t id) {
    for (size_t i = 0; i < q->size; i++) {
        if (q->heap[i]->id == id) {
            event_free(q->heap[i]);
            q->heap[i] = q->heap[q->size - 1];
            q->size--;
            q->heap[q->size] = NULL;
            
            if (i < q->size) {
                size_t parent = (i - 1) / 2;
                if (i > 0 && event_compare(q->heap[i], q->heap[parent]) < 0) {
                    event_queue_heapify_up(q, i);
                } else {
                    event_queue_heapify_down(q, i);
                }
            }
            return 0;
        }
    }
    return -1;
}

event_channel_t* event_channel_create(void) {
    event_channel_t *ch = (event_channel_t*)calloc(1, sizeof(event_channel_t));
    if (!ch) return NULL;
    
    ch->queue = event_queue_create(EVENT_CHANNEL_CAPACITY);
    if (!ch->queue) {
        free(ch);
        return NULL;
    }
    
    pthread_mutex_init(&ch->mutex, NULL);
    pthread_cond_init(&ch->not_empty, NULL);
    pthread_cond_init(&ch->not_full, NULL);
    ch->is_closed = 0;
    
    return ch;
}

void event_channel_destroy(event_channel_t *channel) {
    if (!channel) return;
    
    pthread_mutex_lock(&channel->mutex);
    channel->is_closed = 1;
    pthread_cond_broadcast(&channel->not_empty);
    pthread_cond_broadcast(&channel->not_full);
    pthread_mutex_unlock(&channel->mutex);
    
    event_queue_destroy(channel->queue);
    pthread_mutex_destroy(&channel->mutex);
    pthread_cond_destroy(&channel->not_empty);
    pthread_cond_destroy(&channel->not_full);
    free(channel);
}

int event_channel_post(event_channel_t *channel, event_type_t type,
                        event_priority_t priority, void *data, size_t data_size,
                        event_id_t *out_id) {
    if (!channel) return -EINVAL;
    
    event_t *event = (event_t*)calloc(1, sizeof(event_t));
    if (!event) return -ENOMEM;
    
    event->id = generate_event_id();
    event->type = type;
    event->priority = priority;
    
    if (data && data_size > 0) {
        event->data = malloc(data_size);
        if (!event->data) {
            free(event);
            return -ENOMEM;
        }
        memcpy(event->data, data, data_size);
        event->data_size = data_size;
    } else {
        event->data = NULL;
        event->data_size = 0;
    }
    
    pthread_mutex_lock(&channel->mutex);
    
    if (channel->is_closed) {
        pthread_mutex_unlock(&channel->mutex);
        event_free(event);
        return -EPIPE;
    }
    
    if (channel->queue->size >= channel->queue->capacity) {
        pthread_mutex_unlock(&channel->mutex);
        event_free(event);
        return -EAGAIN;
    }
    
    event_queue_push(channel->queue, event);
    
    if (out_id) {
        *out_id = event->id;
    }
    
    pthread_cond_signal(&channel->not_empty);
    pthread_mutex_unlock(&channel->mutex);
    
    return 0;
}

event_t* event_channel_get(event_channel_t *channel) {
    if (!channel) return NULL;
    
    pthread_mutex_lock(&channel->mutex);
    
    while (channel->queue->size == 0 && !channel->is_closed) {
        pthread_cond_wait(&channel->not_empty, &channel->mutex);
    }
    
    if (channel->is_closed && channel->queue->size == 0) {
        pthread_mutex_unlock(&channel->mutex);
        return NULL;
    }
    
    event_t *event = event_queue_pop(channel->queue);
    
    if (channel->queue->size == channel->queue->capacity - 1) {
        pthread_cond_broadcast(&channel->not_full);
    }
    
    pthread_mutex_unlock(&channel->mutex);
    return event;
}

event_t* event_channel_try_get(event_channel_t *channel) {
    if (!channel) return NULL;
    
    pthread_mutex_lock(&channel->mutex);
    
    if (channel->queue->size == 0) {
        pthread_mutex_unlock(&channel->mutex);
        return NULL;
    }
    
    event_t *event = event_queue_pop(channel->queue);
    
    if (channel->queue->size == channel->queue->capacity - 1) {
        pthread_cond_broadcast(&channel->not_full);
    }
    
    pthread_mutex_unlock(&channel->mutex);
    return event;
}

int event_channel_cancel(event_channel_t *channel, event_id_t id) {
    if (!channel) return -EINVAL;
    
    pthread_mutex_lock(&channel->mutex);
    int result = event_queue_erase(channel->queue, id);
    pthread_mutex_unlock(&channel->mutex);
    
    return result;
}

int event_channel_empty(event_channel_t *channel) {
    if (!channel) return 1;
    
    pthread_mutex_lock(&channel->mutex);
    int empty = (channel->queue->size == 0);
    pthread_mutex_unlock(&channel->mutex);
    
    return empty;
}

size_t event_channel_size(event_channel_t *channel) {
    if (!channel) return 0;
    
    pthread_mutex_lock(&channel->mutex);
    size_t size = channel->queue->size;
    pthread_mutex_unlock(&channel->mutex);
    
    return size;
}

void event_free(event_t *event) {
    if (!event) return;
    if (event->data) {
        free(event->data);
    }
    free(event);
}
