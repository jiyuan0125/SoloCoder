#ifndef EVENT_TIMER_H
#define EVENT_TIMER_H

#include "event_channel.h"
#include <stdint.h>
#include <pthread.h>

typedef uint64_t timer_id_t;

typedef struct timer_event_s {
    timer_id_t timer_id;
    event_id_t event_id;
    event_type_t type;
    event_priority_t priority;
    void *data;
    size_t data_size;
    uint64_t expire_time_ms;
} timer_event_t;

typedef struct timer_heap_s {
    timer_event_t **heap;
    size_t capacity;
    size_t size;
} timer_heap_t;

typedef struct event_timer_s {
    event_channel_t *channel;
    timer_heap_t *heap;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    pthread_t timer_thread;
    volatile int running;
    timer_id_t next_timer_id;
} event_timer_t;

event_timer_t* event_timer_create(event_channel_t *channel);
void event_timer_destroy(event_timer_t *timer);

timer_id_t event_timer_post_delayed(event_timer_t *timer, event_type_t type,
                                      event_priority_t priority, void *data, size_t data_size,
                                      uint64_t delay_ms, event_id_t *out_event_id);

int event_timer_cancel(event_timer_t *timer, timer_id_t timer_id);

#endif
