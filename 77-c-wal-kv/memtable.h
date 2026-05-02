#ifndef MEMTABLE_H
#define MEMTABLE_H

#include "kv_store.h"

struct SkipListNode {
    char* key;
    size_t key_len;
    char* value;
    size_t value_len;
    bool deleted;
    size_t size;
    SkipListNode* forward[];
};

struct SkipList {
    SkipListNode* header;
    int level;
    size_t size;
    size_t count;
};

SkipList* skiplist_create(void);
void skiplist_destroy(SkipList* sl);
int skiplist_put(SkipList* sl, const char* key, size_t key_len, const char* value, size_t value_len);
int skiplist_delete(SkipList* sl, const char* key, size_t key_len);
KVEntry* skiplist_get(SkipList* sl, const char* key, size_t key_len);
SkipListNode* skiplist_first(SkipList* sl);
SkipListNode* skiplist_next(SkipListNode* node);
size_t skiplist_size(SkipList* sl);
size_t skiplist_count(SkipList* sl);

#endif
