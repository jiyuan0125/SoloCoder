#ifndef KV_STORE_H
#define KV_STORE_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define MEMTABLE_MAX_SIZE (4 * 1024 * 1024)
#define MAX_KEY_LEN 65535
#define MAX_VALUE_LEN (1024 * 1024)
#define SKIPLIST_MAX_LEVEL 12
#define COMPACTION_INTERVAL_SECONDS 30

typedef enum {
    RECORD_PUT = 0,
    RECORD_DELETE = 1
} RecordType;

typedef struct {
    char* key;
    size_t key_len;
    char* value;
    size_t value_len;
    bool deleted;
} KVEntry;

typedef struct SkipListNode SkipListNode;
typedef struct SkipList SkipList;
typedef struct WAL WAL;
typedef struct SSTable SSTable;
typedef struct SSTableList SSTableList;
typedef struct CompactionTask CompactionTask;
typedef struct DB DB;

void kv_entry_free(KVEntry* entry);
KVEntry* kv_entry_create(const char* key, size_t key_len, const char* value, size_t value_len, bool deleted);
KVEntry* kv_entry_clone(const KVEntry* entry);

uint32_t crc32(const uint8_t* data, size_t len);

#endif
