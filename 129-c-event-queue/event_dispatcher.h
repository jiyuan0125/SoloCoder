#ifndef EVENT_DISPATCHER_H
#define EVENT_DISPATCHER_H

#include "event_channel.h"

#define MAX_EVENT_TYPES 256

typedef void (*event_handler_fn)(event_t *event, void *user_data);

typedef struct handler_entry_s {
    event_handler_fn handler;
    void *user_data;
} handler_entry_t;

typedef struct event_dispatcher_s {
    event_channel_t *channel;
    handler_entry_t handlers[MAX_EVENT_TYPES];
    pthread_mutex_t handler_mutex;
    volatile int running;
} event_dispatcher_t;

event_dispatcher_t* event_dispatcher_create(event_channel_t *channel);
void event_dispatcher_destroy(event_dispatcher_t *dispatcher);

int event_dispatcher_register(event_dispatcher_t *dispatcher, event_type_t type,
                                event_handler_fn handler, void *user_data);

int event_dispatcher_unregister(event_dispatcher_t *dispatcher, event_type_t type);

int event_dispatcher_dispatch(event_dispatcher_t *dispatcher, event_t *event);

void event_dispatcher_run(event_dispatcher_t *dispatcher);
void event_dispatcher_stop(event_dispatcher_t *dispatcher);

#endif
