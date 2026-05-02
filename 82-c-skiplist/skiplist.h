#ifndef SKIPLIST_H
#define SKIPLIST_H

#define _XOPEN_SOURCE 700
#include <pthread.h>
#include <stdatomic.h>
#include <stdio.h>

#define MAX_LEVEL 16

typedef struct skiplist_node {
    int key;
    void *value;
    int level;
    struct skiplist_node *forward[];
} skiplist_node_t;

typedef struct skiplist {
    skiplist_node_t *head;
    pthread_rwlock_t rwlock;
    atomic_ulong count;
    int max_level;
} skiplist_t;

typedef void (*range_callback)(int key, void *value);

skiplist_t *skiplist_create(void);
void skiplist_destroy(skiplist_t *sl);
int skiplist_insert(skiplist_t *sl, int key, void *value);
void *skiplist_search(skiplist_t *sl, int key);
int skiplist_delete(skiplist_t *sl, int key);
void skiplist_range_query(skiplist_t *sl, int low, int high, range_callback callback);
unsigned long skiplist_count(skiplist_t *sl);
void skiplist_level_stats(skiplist_t *sl);

#endif
