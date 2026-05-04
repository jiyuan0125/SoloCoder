#ifndef LOCK_FREE_DEQUE_H
#define LOCK_FREE_DEQUE_H

#include <stdint.h>
#include <stdbool.h>
#include <pthread.h>
#include <stddef.h>

typedef struct lf_node_s {
    struct lf_node_s* next;
    struct lf_node_s* prev;
    void* data;
} lf_node_t;

typedef struct lf_deque_s {
    lf_node_t* head;
    lf_node_t* tail;
    pthread_spinlock_t lock;
    size_t size;
    size_t max_size;
} lf_deque_t;

typedef enum {
    LF_DEQUE_OK = 0,
    LF_DEQUE_FULL = -1,
    LF_DEQUE_EMPTY = -2,
    LF_DEQUE_ERROR = -3
} lf_deque_status_t;

typedef void (*lf_data_destructor_t)(void* data);

lf_deque_t* lf_deque_create(size_t max_size);
void lf_deque_destroy(lf_deque_t* deque, lf_data_destructor_t destructor);

lf_deque_status_t lf_deque_push_back(lf_deque_t* deque, void* data);
lf_deque_status_t lf_deque_push_front(lf_deque_t* deque, void* data);

void* lf_deque_pop_back(lf_deque_t* deque);
void* lf_deque_pop_front(lf_deque_t* deque);

void* lf_deque_peek_front(lf_deque_t* deque);
void* lf_deque_peek_back(lf_deque_t* deque);

size_t lf_deque_size(lf_deque_t* deque);
bool lf_deque_empty(lf_deque_t* deque);
bool lf_deque_full(lf_deque_t* deque);

size_t lf_deque_drain(lf_deque_t* deque, void*** out_array, size_t* out_count);
void lf_deque_drop_oldest(lf_deque_t* deque, size_t count, lf_data_destructor_t destructor);

#endif
