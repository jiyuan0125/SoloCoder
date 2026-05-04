#ifndef EVENT_CHANNEL_H
#define EVENT_CHANNEL_H

#include <stdint.h>
#include <stddef.h>
#include <pthread.h>

#define EVENT_CHANNEL_CAPACITY 1024

typedef uint64_t event_id_t;
typedef uint32_t event_type_t;
typedef uint32_t event_priority_t;

#define EVENT_PRIORITY_HIGHEST  0
#define EVENT_PRIORITY_HIGH     1
#define EVENT_PRIORITY_NORMAL   2
#define EVENT_PRIORITY_LOW      3
#define EVENT_PRIORITY_LOWEST   4

typedef struct event_s {
    event_id_t id;
    event_type_t type;
    event_priority_t priority;
    void *data;
    size_t data_size;
    uint64_t sequence;
} event_t;

typedef struct event_queue_s {
    event_t **heap;
    size_t capacity;
    size_t size;
    uint64_t sequence_counter;
} event_queue_t;

typedef struct event_channel_s {
    event_queue_t *queue;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
    volatile int is_closed;
} event_channel_t;

event_channel_t* event_channel_create(void);
void event_channel_destroy(event_channel_t *channel);

int event_channel_post(event_channel_t *channel, event_type_t type, 
                        event_priority_t priority, void *data, size_t data_size,
                        event_id_t *out_id);

event_t* event_channel_get(event_channel_t *channel);
event_t* event_channel_try_get(event_channel_t *channel);

int event_channel_cancel(event_channel_t *channel, event_id_t id);

int event_channel_empty(event_channel_t *channel);
size_t event_channel_size(event_channel_t *channel);

void event_free(event_t *event);

#endif
