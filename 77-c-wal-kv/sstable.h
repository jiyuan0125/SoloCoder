#ifndef SSTABLE_H
#define SSTABLE_H

#include "kv_store.h"
#include <stdio.h>

struct SSTable {
    char* path;
    FILE* file;
    uint64_t id;
    size_t entry_count;
};

struct SSTableList {
    SSTable** tables;
    size_t count;
    size_t capacity;
};

SSTableList* sstable_list_create(void);
void sstable_list_destroy(SSTableList* list);
int sstable_list_add(SSTableList* list, SSTable* table);
void sstable_list_remove(SSTableList* list, size_t index);
SSTable* sstable_list_get(SSTableList* list, size_t index);
size_t sstable_list_count(SSTableList* list);

SSTable* sstable_create_from_memtable(const char* dir, uint64_t id, SkipList* memtable);
SSTable* sstable_open(const char* path);
void sstable_close(SSTable* table);
KVEntry* sstable_get(SSTable* table, const char* key, size_t key_len);
uint64_t sstable_get_id(SSTable* table);

#endif
