#include "db.h"
#include "wal.h"
#include "memtable.h"
#include "sstable.h"
#include "compaction.h"
#include "kv_store.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <sys/stat.h>
#include <dirent.h>
#include <errno.h>

static int key_compare(const char* key1, size_t len1, const char* key2, size_t len2) {
    size_t min_len = len1 < len2 ? len1 : len2;
    int cmp = memcmp(key1, key2, min_len);
    if (cmp != 0) return cmp;
    if (len1 < len2) return -1;
    if (len1 > len2) return 1;
    return 0;
}

static int key_starts_with(const char* key, size_t key_len, const char* prefix, size_t prefix_len) {
    if (key_len < prefix_len) return 0;
    return memcmp(key, prefix, prefix_len) == 0;
}

static int ensure_dir_exists(const char* path) {
    struct stat st;
    if (stat(path, &st) == 0) {
        if (S_ISDIR(st.st_mode)) return 0;
        return -1;
    }
    
    #ifdef _WIN32
    return mkdir(path);
    #else
    return mkdir(path, 0755);
    #endif
}

static uint64_t extract_sstable_id(const char* filename) {
    uint64_t id = 0;
    if (sscanf(filename, "sst_%lu.sst", (unsigned long*)&id) == 1) {
        return id;
    }
    return 0;
}

static int flush_memtable(DB* db) {
    if (!db || !db->memtable) return -1;
    
    if (skiplist_count(db->memtable) == 0) return 0;
    
    uint64_t sstable_id = db->next_sstable_id++;
    
    SSTable* new_table = sstable_create_from_memtable(db->path, sstable_id, db->memtable);
    if (!new_table) return -1;
    
    sstable_list_add(db->sstables, new_table);
    
    skiplist_destroy(db->memtable);
    db->memtable = skiplist_create();
    if (!db->memtable) {
        return -1;
    }
    
    char old_wal_path[512];
    snprintf(old_wal_path, sizeof(old_wal_path), "%s/wal.log", db->path);
    
    if (db->wal) {
        wal_close(db->wal);
        db->wal = NULL;
    }
    
    remove(old_wal_path);
    
    db->wal = wal_open(old_wal_path);
    if (!db->wal) {
        return -1;
    }
    
    return 0;
}

DB* db_open(const char* path) {
    if (!path) return NULL;
    
    DB* db = (DB*)malloc(sizeof(DB));
    if (!db) return NULL;
    memset(db, 0, sizeof(DB));
    
    db->path = strdup(path);
    if (!db->path) {
        free(db);
        return NULL;
    }
    
    if (ensure_dir_exists(db->path) != 0) {
        free(db->path);
        free(db);
        return NULL;
    }
    
    if (pthread_rwlock_init(&db->rwlock, NULL) != 0) {
        free(db->path);
        free(db);
        return NULL;
    }
    
    db->memtable = skiplist_create();
    if (!db->memtable) {
        pthread_rwlock_destroy(&db->rwlock);
        free(db->path);
        free(db);
        return NULL;
    }
    
    db->sstables = sstable_list_create();
    if (!db->sstables) {
        skiplist_destroy(db->memtable);
        pthread_rwlock_destroy(&db->rwlock);
        free(db->path);
        free(db);
        return NULL;
    }
    
    db->next_sstable_id = 1;
    
    DIR* dir = opendir(db->path);
    if (dir) {
        struct dirent* entry;
        while ((entry = readdir(dir)) != NULL) {
            uint64_t id = extract_sstable_id(entry->d_name);
            if (id > 0) {
                char sst_path[512];
                snprintf(sst_path, sizeof(sst_path), "%s/%s", db->path, entry->d_name);
                SSTable* table = sstable_open(sst_path);
                if (table) {
                    sstable_list_add(db->sstables, table);
                    if (id >= db->next_sstable_id) {
                        db->next_sstable_id = id + 1;
                    }
                }
            }
        }
        closedir(dir);
    }
    
    char wal_path[512];
    snprintf(wal_path, sizeof(wal_path), "%s/wal.log", db->path);
    db->wal = wal_open(wal_path);
    if (!db->wal) {
        sstable_list_destroy(db->sstables);
        skiplist_destroy(db->memtable);
        pthread_rwlock_destroy(&db->rwlock);
        free(db->path);
        free(db);
        return NULL;
    }
    
    wal_replay(db->wal, db->memtable);
    
    db->compaction = compaction_task_create(db);
    if (db->compaction) {
        compaction_task_start(db->compaction);
    }
    
    db->closed = false;
    return db;
}

void db_close(DB* db) {
    if (!db) return;
    
    db->closed = true;
    
    if (db->compaction) {
        compaction_task_destroy(db->compaction);
        db->compaction = NULL;
    }
    
    pthread_rwlock_wrlock(&db->rwlock);
    
    if (skiplist_count(db->memtable) > 0) {
        flush_memtable(db);
    }
    
    if (db->wal) {
        wal_sync(db->wal);
        wal_close(db->wal);
        db->wal = NULL;
    }
    
    if (db->memtable) {
        skiplist_destroy(db->memtable);
        db->memtable = NULL;
    }
    
    if (db->sstables) {
        sstable_list_destroy(db->sstables);
        db->sstables = NULL;
    }
    
    pthread_rwlock_unlock(&db->rwlock);
    pthread_rwlock_destroy(&db->rwlock);
    
    free(db->path);
    free(db);
}

int db_put(DB* db, const char* key, size_t key_len, const char* value, size_t value_len) {
    if (!db || !key || key_len == 0 || key_len > MAX_KEY_LEN) return -1;
    if (value_len > MAX_VALUE_LEN) return -1;
    
    pthread_rwlock_wrlock(&db->rwlock);
    
    int ret = wal_append(db->wal, RECORD_PUT, key, key_len, value, value_len);
    if (ret != 0) {
        pthread_rwlock_unlock(&db->rwlock);
        return ret;
    }
    
    ret = skiplist_put(db->memtable, key, key_len, value, value_len);
    if (ret != 0) {
        pthread_rwlock_unlock(&db->rwlock);
        return ret;
    }
    
    wal_sync(db->wal);
    
    if (skiplist_size(db->memtable) >= MEMTABLE_MAX_SIZE) {
        ret = flush_memtable(db);
    }
    
    pthread_rwlock_unlock(&db->rwlock);
    return ret;
}

int db_delete(DB* db, const char* key, size_t key_len) {
    if (!db || !key || key_len == 0 || key_len > MAX_KEY_LEN) return -1;
    
    pthread_rwlock_wrlock(&db->rwlock);
    
    int ret = wal_append(db->wal, RECORD_DELETE, key, key_len, NULL, 0);
    if (ret != 0) {
        pthread_rwlock_unlock(&db->rwlock);
        return ret;
    }
    
    ret = skiplist_delete(db->memtable, key, key_len);
    wal_sync(db->wal);
    
    pthread_rwlock_unlock(&db->rwlock);
    return ret;
}

KVEntry* db_get(DB* db, const char* key, size_t key_len) {
    if (!db || !key || key_len == 0) return NULL;
    
    pthread_rwlock_rdlock(&db->rwlock);
    
    KVEntry* entry = skiplist_get(db->memtable, key, key_len);
    if (entry) {
        pthread_rwlock_unlock(&db->rwlock);
        return entry;
    }
    
    size_t count = sstable_list_count(db->sstables);
    for (size_t i = count; i > 0; i--) {
        SSTable* table = sstable_list_get(db->sstables, i - 1);
        if (table) {
            KVEntry* sst_entry = sstable_get(table, key, key_len);
            if (sst_entry) {
                pthread_rwlock_unlock(&db->rwlock);
                return sst_entry;
            }
        }
    }
    
    pthread_rwlock_unlock(&db->rwlock);
    return NULL;
}

static int key_in_results(KVEntry** results, size_t count, const char* key, size_t key_len) {
    for (size_t i = 0; i < count; i++) {
        if (key_compare(results[i]->key, results[i]->key_len, key, key_len) == 0) {
            return 1;
        }
    }
    return 0;
}

static int entry_compare(const void* a, const void* b) {
    KVEntry* entry_a = *(KVEntry**)a;
    KVEntry* entry_b = *(KVEntry**)b;
    return key_compare(entry_a->key, entry_a->key_len, entry_b->key, entry_b->key_len);
}

KVEntry** db_scan(DB* db, const char* prefix, size_t prefix_len, size_t* out_count) {
    if (!db || !prefix || prefix_len == 0 || !out_count) {
        if (out_count) *out_count = 0;
        return NULL;
    }
    
    pthread_rwlock_rdlock(&db->rwlock);
    
    size_t result_count = 0;
    size_t result_cap = 16;
    KVEntry** results = (KVEntry**)malloc(result_cap * sizeof(KVEntry*));
    if (!results) {
        pthread_rwlock_unlock(&db->rwlock);
        *out_count = 0;
        return NULL;
    }
    
    SkipListNode* node = skiplist_first(db->memtable);
    while (node) {
        if (key_starts_with(node->key, node->key_len, prefix, prefix_len)) {
            if (!node->deleted) {
                if (result_count >= result_cap) {
                    size_t new_cap = result_cap * 2;
                    KVEntry** new_results = (KVEntry**)realloc(results, new_cap * sizeof(KVEntry*));
                    if (!new_results) {
                        for (size_t i = 0; i < result_count; i++) {
                            kv_entry_free(results[i]);
                        }
                        free(results);
                        pthread_rwlock_unlock(&db->rwlock);
                        *out_count = 0;
                        return NULL;
                    }
                    results = new_results;
                    result_cap = new_cap;
                }
                results[result_count++] = kv_entry_create(node->key, node->key_len, 
                                                           node->value, node->value_len, false);
            }
        }
        node = skiplist_next(node);
    }
    
    size_t sstable_count = sstable_list_count(db->sstables);
    for (size_t i = sstable_count; i > 0; i--) {
        SSTable* table = sstable_list_get(db->sstables, i - 1);
        if (!table || !table->file) continue;
        
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
            
            int cmp = key_compare(key, key_len, prefix, prefix_len);
            if (cmp < 0) {
                uint32_t value_len;
                if (fread(&value_len, sizeof(uint32_t), 1, table->file) != 1) {
                    free(key);
                    break;
                }
                fseek(table->file, value_len, SEEK_CUR);
                free(key);
                continue;
            }
            
            if (!key_starts_with(key, key_len, prefix, prefix_len)) {
                free(key);
                break;
            }
            
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
            
            if (!key_in_results(results, result_count, key, key_len)) {
                if (result_count >= result_cap) {
                    size_t new_cap = result_cap * 2;
                    KVEntry** new_results = (KVEntry**)realloc(results, new_cap * sizeof(KVEntry*));
                    if (!new_results) {
                        free(key);
                        free(value);
                        for (size_t j = 0; j < result_count; j++) {
                            kv_entry_free(results[j]);
                        }
                        free(results);
                        pthread_rwlock_unlock(&db->rwlock);
                        *out_count = 0;
                        return NULL;
                    }
                    results = new_results;
                    result_cap = new_cap;
                }
                results[result_count++] = kv_entry_create(key, key_len, value, value_len, false);
            }
            
            free(key);
            free(value);
        }
    }
    
    if (result_count > 1) {
        qsort(results, result_count, sizeof(KVEntry*), entry_compare);
    }
    
    pthread_rwlock_unlock(&db->rwlock);
    
    *out_count = result_count;
    return results;
}

void db_scan_results_free(KVEntry** results, size_t count) {
    if (!results) return;
    for (size_t i = 0; i < count; i++) {
        kv_entry_free(results[i]);
    }
    free(results);
}
