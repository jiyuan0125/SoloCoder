#ifndef DB_H
#define DB_H

#include "kv_store.h"
#include "memtable.h"
#include "sstable.h"
#include "wal.h"
#include "compaction.h"
#include <pthread.h>

struct DB {
    char* path;
    SkipList* memtable;
    SSTableList* sstables;
    WAL* wal;
    uint64_t next_sstable_id;
    pthread_rwlock_t rwlock;
    CompactionTask* compaction;
    bool closed;
};

DB* db_open(const char* path);
void db_close(DB* db);
int db_put(DB* db, const char* key, size_t key_len, const char* value, size_t value_len);
int db_delete(DB* db, const char* key, size_t key_len);
KVEntry* db_get(DB* db, const char* key, size_t key_len);
KVEntry** db_scan(DB* db, const char* prefix, size_t prefix_len, size_t* out_count);
void db_scan_results_free(KVEntry** results, size_t count);

#endif
