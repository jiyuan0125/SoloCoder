#ifndef BTREE_COMMON_H
#define BTREE_COMMON_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <pthread.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <time.h>

#define PAGE_SIZE 4096
#define MAX_KEY_SIZE 256
#define MAX_VALUE_SIZE 4096
#define CACHE_CAPACITY 1000

#define INVALID_PAGE_NUM UINT32_MAX
#define HEADER_PAGE_NUM 0
#define ROOT_PAGE_INITIAL 1

typedef uint32_t page_num_t;

#define PAGE_TYPE_INTERNAL 1
#define PAGE_TYPE_LEAF 2
#define PAGE_TYPE_VALUE 3

typedef uint8_t page_type_t;

typedef struct {
    uint16_t key_len;
    uint8_t key_data[MAX_KEY_SIZE];
} bt_key_t;

typedef struct {
    uint32_t value_len;
    uint8_t value_data[MAX_VALUE_SIZE];
} bt_value_t;

typedef struct {
    uint32_t magic;
    uint32_t version;
    page_num_t root_page;
    page_num_t num_pages;
    uint32_t freelist_head;
    uint32_t reserved[1010];
} file_header_t;

typedef struct {
    page_num_t page_num;
    page_type_t type;
    uint8_t reserved1;
    uint16_t num_entries;
    uint16_t reserved2;
    page_num_t next_leaf;
    page_num_t prev_leaf;
    uint8_t data[PAGE_SIZE - 20];
} page_t;

typedef struct {
    bt_key_t key;
    page_num_t child;
} internal_entry_t;

typedef struct {
    bt_key_t key;
    page_num_t value_page;
    uint32_t value_len;
} leaf_entry_t;

typedef struct {
    uint8_t data[PAGE_SIZE - 16];
} value_page_t;

typedef struct cache_entry {
    page_num_t page_num;
    page_t page;
    int is_dirty;
    int ref_count;
    time_t last_access;
    struct cache_entry *prev;
    struct cache_entry *next;
    struct cache_entry *lru_prev;
    struct cache_entry *lru_next;
} cache_entry_t;

typedef struct {
    cache_entry_t *hash_table[10007];
    cache_entry_t *lru_head;
    cache_entry_t *lru_tail;
    int size;
    int capacity;
    pthread_mutex_t mutex;
} page_cache_t;

typedef enum {
    WAL_OP_TYPE_INSERT = 1,
    WAL_OP_TYPE_DELETE = 2,
    WAL_OP_TYPE_UPDATE = 3,
    WAL_OP_TYPE_SPLIT = 4,
    WAL_OP_TYPE_COMMIT = 5,
    WAL_OP_TYPE_ABORT = 6
} wal_op_type_t;

typedef struct {
    uint32_t lsn;
    uint32_t prev_lsn;
    uint32_t op_type;
    page_num_t page_num;
    uint16_t key_len;
    uint32_t value_len;
    uint16_t value_offset;
} wal_header_t;

typedef struct {
    int fd;
    uint32_t next_lsn;
    uint32_t last_lsn;
    pthread_mutex_t mutex;
} wal_t;

typedef struct {
    int data_fd;
    file_header_t header;
    page_cache_t cache;
    wal_t wal;
    pthread_rwlock_t tree_lock;
    int is_open;
} btree_t;

static inline int key_compare(const bt_key_t *a, const bt_key_t *b) {
    if (a->key_len != b->key_len) {
        return (a->key_len < b->key_len) ? -1 : 1;
    }
    return memcmp(a->key_data, b->key_data, a->key_len);
}

static inline int key_copy(bt_key_t *dest, const bt_key_t *src) {
    if (src->key_len > MAX_KEY_SIZE) {
        return -1;
    }
    dest->key_len = src->key_len;
    memcpy(dest->key_data, src->key_data, src->key_len);
    return 0;
}

static inline int value_copy(bt_value_t *dest, const bt_value_t *src) {
    if (src->value_len > MAX_VALUE_SIZE) {
        return -1;
    }
    dest->value_len = src->value_len;
    memcpy(dest->value_data, src->value_data, src->value_len);
    return 0;
}

#endif
