#ifndef WAL_H
#define WAL_H

#include "kv_store.h"
#include <stdio.h>

struct WAL {
    char* path;
    FILE* file;
};

WAL* wal_open(const char* path);
void wal_close(WAL* wal);
int wal_append(WAL* wal, RecordType type, const char* key, size_t key_len, const char* value, size_t value_len);
int wal_replay(WAL* wal, SkipList* memtable);
void wal_sync(WAL* wal);

#endif
