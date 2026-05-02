#ifndef EVENT_LOOP_H
#define EVENT_LOOP_H

#include <stdint.h>
#include <sys/epoll.h>

typedef void (*event_callback_t)(int events, void *data);
typedef uint64_t timer_id_t;

typedef struct event_loop event_loop_t;
typedef struct timer_node timer_node_t;

struct timer_node {
    timer_id_t id;
    uint64_t expire_ms;
    uint64_t interval_ms;
    int is_repeat;
    event_callback_t callback;
    void *data;
    int heap_index;
    int is_pending;
};

event_loop_t* loop_create(void);
void loop_destroy(event_loop_t *loop);

int loop_add_fd(event_loop_t *loop, int fd, int events, event_callback_t callback, void *data);
int loop_remove_fd(event_loop_t *loop, int fd);
int loop_modify_fd(event_loop_t *loop, int fd, int events);

timer_id_t loop_add_timer(event_loop_t *loop, uint64_t delay_ms, event_callback_t callback, void *data, int is_repeat);
int loop_remove_timer(event_loop_t *loop, timer_id_t timer_id);

uint64_t loop_run(event_loop_t *loop);
void loop_stop(event_loop_t *loop);

#endif
