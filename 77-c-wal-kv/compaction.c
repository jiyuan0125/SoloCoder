#include "compaction.h"
#include "db.h"
#include "sstable.h"
#include "memtable.h"
#include "kv_store.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <unistd.h>
#include <errno.h>

typedef struct {
    char* key;
    size_t key_len;
    char* value;
    size_t value_len;
    bool deleted;
} MergeEntry;

static MergeEntry* merge_entry_create(const char* key, size_t key_len, const char* value, size_t value_len, bool deleted) {
    MergeEntry* entry = (MergeEntry*)malloc(sizeof(MergeEntry));
    if (!entry) return NULL;
    
    entry->key = (char*)malloc(key_len + 1);
    if (!entry->key) {
        free(entry);
        return NULL;
    }
    memcpy(entry->key, key, key_len);
    entry->key[key_len] = '\0';
    entry->key_len = key_len;
    
    if (value && value_len > 0) {
        entry->value = (char*)malloc(value_len + 1);
        if (!entry->value) {
            free(entry->key);
            free(entry);
            return NULL;
        }
        memcpy(entry->value, value, value_len);
        entry->value[value_len] = '\0';
        entry->value_len = value_len;
    } else {
        entry->value = NULL;
        entry->value_len = 0;
    }
    
    entry->deleted = deleted;
    return entry;
}

static void merge_entry_free(MergeEntry* entry) {
    if (!entry) return;
    free(entry->key);
    free(entry->value);
    free(entry);
}

static int key_compare(const char* key1, size_t len1, const char* key2, size_t len2) {
    size_t min_len = len1 < len2 ? len1 : len2;
    int cmp = memcmp(key1, key2, min_len);
    if (cmp != 0) return cmp;
    if (len1 < len2) return -1;
    if (len1 > len2) return 1;
    return 0;
}

static SkipList* load_sstable_to_memtable(SSTable* table) {
    if (!table || !table->file) return NULL;
    
    SkipList* sl = skiplist_create();
    if (!sl) return NULL;
    
    fseek(table->file, 0, SEEK_SET);
    
    while (1) {
        uint32_t key_len;
        if (fread(&key_len, sizeof(uint32_t), 1, table->file) != 1) break;
        
        char* key = (char*)malloc(key_len + 1);
        if (!key) break;
        
        if (fread(key, 1, key_len, table->file) != key_len) {
            free(key);
            break;
        }
        key[key_len] = '\0';
        
        uint32_t value_len;
        if (fread(&value_len, sizeof(uint32_t), 1, table->file) != 1) {
            free(key);
            break;
        }
        
        char* value = NULL;
        if (value_len > 0) {
            value = (char*)malloc(value_len + 1);
            if (!value) {
                free(key);
                break;
            }
            if (fread(value, 1, value_len, table->file) != value_len) {
                free(value);
                free(key);
                break;
            }
            value[value_len] = '\0';
        }
        
        skiplist_put(sl, key, key_len, value, value_len);
        
        free(key);
        free(value);
    }
    
    return sl;
}

static int perform_compaction(DB* db) {
    if (!db) return -1;
    
    pthread_rwlock_rdlock(&db->rwlock);
    
    size_t sstable_count = sstable_list_count(db->sstables);
    if (sstable_count < 2) {
        pthread_rwlock_unlock(&db->rwlock);
        return 0;
    }
    
    size_t compact_count = sstable_count > 4 ? 4 : sstable_count;
    SSTable** tables_to_compact = (SSTable**)malloc(compact_count * sizeof(SSTable*));
    if (!tables_to_compact) {
        pthread_rwlock_unlock(&db->rwlock);
        return -1;
    }
    
    for (size_t i = 0; i < compact_count; i++) {
        tables_to_compact[i] = sstable_list_get(db->sstables, i);
    }
    
    pthread_rwlock_unlock(&db->rwlock);
    
    SkipList* merged = skiplist_create();
    if (!merged) {
        free(tables_to_compact);
        return -1;
    }
    
    for (size_t i = 0; i < compact_count; i++) {
        SkipList* sl = load_sstable_to_memtable(tables_to_compact[i]);
        if (sl) {
            SkipListNode* node = skiplist_first(sl);
            while (node) {
                skiplist_put(merged, node->key, node->key_len, node->value, node->value_len);
                node = skiplist_next(node);
            }
            skiplist_destroy(sl);
        }
    }
    
    pthread_rwlock_wrlock(&db->rwlock);
    uint64_t new_id = db->next_sstable_id++;
    pthread_rwlock_unlock(&db->rwlock);
    
    SSTable* new_table = sstable_create_from_memtable(db->path, new_id, merged);
    skiplist_destroy(merged);
    
    if (!new_table) {
        free(tables_to_compact);
        return -1;
    }
    
    pthread_rwlock_wrlock(&db->rwlock);
    
    for (size_t i = 0; i < compact_count; i++) {
        SSTable* old = sstable_list_get(db->sstables, 0);
        if (old) {
            remove(old->path);
        }
        sstable_list_remove(db->sstables, 0);
    }
    
    sstable_list_add(db->sstables, new_table);
    
    pthread_rwlock_unlock(&db->rwlock);
    
    free(tables_to_compact);
    return 0;
}

static void* compaction_thread(void* arg) {
    CompactionTask* task = (CompactionTask*)arg;
    if (!task) return NULL;
    
    pthread_mutex_lock(&task->mutex);
    task->running = true;
    pthread_mutex_unlock(&task->mutex);
    
    while (1) {
        pthread_mutex_lock(&task->mutex);
        
        if (task->stop_requested) {
            task->running = false;
            pthread_mutex_unlock(&task->mutex);
            break;
        }
        
        struct timespec ts;
        clock_gettime(CLOCK_REALTIME, &ts);
        ts.tv_sec += COMPACTION_INTERVAL_SECONDS;
        
        int ret = pthread_cond_timedwait(&task->cond, &task->mutex, &ts);
        
        if (task->stop_requested) {
            task->running = false;
            pthread_mutex_unlock(&task->mutex);
            break;
        }
        
        pthread_mutex_unlock(&task->mutex);
        
        if (task->db && !task->db->closed) {
            perform_compaction(task->db);
        }
    }
    
    return NULL;
}

CompactionTask* compaction_task_create(DB* db) {
    CompactionTask* task = (CompactionTask*)malloc(sizeof(CompactionTask));
    if (!task) return NULL;
    
    task->db = db;
    task->running = false;
    task->stop_requested = false;
    
    if (pthread_mutex_init(&task->mutex, NULL) != 0) {
        free(task);
        return NULL;
    }
    
    if (pthread_cond_init(&task->cond, NULL) != 0) {
        pthread_mutex_destroy(&task->mutex);
        free(task);
        return NULL;
    }
    
    return task;
}

void compaction_task_destroy(CompactionTask* task) {
    if (!task) return;
    
    compaction_task_stop(task);
    
    pthread_mutex_destroy(&task->mutex);
    pthread_cond_destroy(&task->cond);
    free(task);
}

int compaction_task_start(CompactionTask* task) {
    if (!task) return -1;
    
    pthread_mutex_lock(&task->mutex);
    if (task->running) {
        pthread_mutex_unlock(&task->mutex);
        return 0;
    }
    task->stop_requested = false;
    pthread_mutex_unlock(&task->mutex);
    
    if (pthread_create(&task->thread, NULL, compaction_thread, task) != 0) {
        return -1;
    }
    
    return 0;
}

void compaction_task_stop(CompactionTask* task) {
    if (!task) return;
    
    pthread_mutex_lock(&task->mutex);
    if (!task->running) {
        pthread_mutex_unlock(&task->mutex);
        return;
    }
    
    task->stop_requested = true;
    pthread_cond_signal(&task->cond);
    pthread_mutex_unlock(&task->mutex);
    
    pthread_join(task->thread, NULL);
}

void compaction_trigger(CompactionTask* task) {
    if (!task) return;
    
    pthread_mutex_lock(&task->mutex);
    pthread_cond_signal(&task->cond);
    pthread_mutex_unlock(&task->mutex);
}
