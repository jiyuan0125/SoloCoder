#ifndef EVENT_LOOP_INTERNAL_H
#define EVENT_LOOP_INTERNAL_H

#include <stdint.h>
#include "event_loop.h"

#define INITIAL_HEAP_CAPACITY 64

struct timer_heap {
    timer_node_t **nodes;
    int size;
    int capacity;
};

struct timer_heap* timer_heap_create(void);
void timer_heap_destroy(struct timer_heap *heap);
int timer_heap_insert(struct timer_heap *heap, timer_node_t *node);
timer_node_t* timer_heap_peek(struct timer_heap *heap);
timer_node_t* timer_heap_pop(struct timer_heap *heap);
int timer_heap_remove(struct timer_heap *heap, timer_node_t *node);

#endif
