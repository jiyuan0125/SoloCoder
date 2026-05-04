#include "event_dispatcher.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>

event_dispatcher_t* event_dispatcher_create(event_channel_t *channel) {
    if (!channel) return NULL;
    
    event_dispatcher_t *d = (event_dispatcher_t*)calloc(1, sizeof(event_dispatcher_t));
    if (!d) return NULL;
    
    d->channel = channel;
    memset(d->handlers, 0, sizeof(d->handlers));
    pthread_mutex_init(&d->handler_mutex, NULL);
    d->running = 0;
    
    return d;
}

void event_dispatcher_destroy(event_dispatcher_t *dispatcher) {
    if (!dispatcher) return;
    
    pthread_mutex_destroy(&dispatcher->handler_mutex);
    free(dispatcher);
}

int event_dispatcher_register(event_dispatcher_t *dispatcher, event_type_t type,
                                event_handler_fn handler, void *user_data) {
    if (!dispatcher || !handler) return -EINVAL;
    if (type >= MAX_EVENT_TYPES) return -EINVAL;
    
    pthread_mutex_lock(&dispatcher->handler_mutex);
    
    dispatcher->handlers[type].handler = handler;
    dispatcher->handlers[type].user_data = user_data;
    
    pthread_mutex_unlock(&dispatcher->handler_mutex);
    
    return 0;
}

int event_dispatcher_unregister(event_dispatcher_t *dispatcher, event_type_t type) {
    if (!dispatcher) return -EINVAL;
    if (type >= MAX_EVENT_TYPES) return -EINVAL;
    
    pthread_mutex_lock(&dispatcher->handler_mutex);
    
    dispatcher->handlers[type].handler = NULL;
    dispatcher->handlers[type].user_data = NULL;
    
    pthread_mutex_unlock(&dispatcher->handler_mutex);
    
    return 0;
}

int event_dispatcher_dispatch(event_dispatcher_t *dispatcher, event_t *event) {
    if (!dispatcher || !event) return -EINVAL;
    if (event->type >= MAX_EVENT_TYPES) return -ENOENT;
    
    event_handler_fn handler = NULL;
    void *user_data = NULL;
    
    pthread_mutex_lock(&dispatcher->handler_mutex);
    
    if (dispatcher->handlers[event->type].handler) {
        handler = dispatcher->handlers[event->type].handler;
        user_data = dispatcher->handlers[event->type].user_data;
    }
    
    pthread_mutex_unlock(&dispatcher->handler_mutex);
    
    if (handler) {
        handler(event, user_data);
        return 0;
    }
    
    return -ENOENT;
}

void event_dispatcher_run(event_dispatcher_t *dispatcher) {
    if (!dispatcher || !dispatcher->channel) return;
    
    dispatcher->running = 1;
    
    while (dispatcher->running) {
        event_t *event = event_channel_get(dispatcher->channel);
        if (!event) {
            break;
        }
        
        event_dispatcher_dispatch(dispatcher, event);
        event_free(event);
    }
}

void event_dispatcher_stop(event_dispatcher_t *dispatcher) {
    if (!dispatcher) return;
    
    dispatcher->running = 0;
}
