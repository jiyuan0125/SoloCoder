#include "cache.h"

#define HASH_TABLE_SIZE 10007

static uint32_t hash_func(page_num_t page_num) {
    return (page_num * 2654435761U) % HASH_TABLE_SIZE;
}

static void lru_remove(page_cache_t *cache, cache_entry_t *entry) {
    if (entry->lru_prev) {
        entry->lru_prev->lru_next = entry->lru_next;
    } else {
        cache->lru_head = entry->lru_next;
    }
    if (entry->lru_next) {
        entry->lru_next->lru_prev = entry->lru_prev;
    } else {
        cache->lru_tail = entry->lru_prev;
    }
    entry->lru_prev = NULL;
    entry->lru_next = NULL;
}

static void lru_add_to_head(page_cache_t *cache, cache_entry_t *entry) {
    entry->lru_prev = NULL;
    entry->lru_next = cache->lru_head;
    if (cache->lru_head) {
        cache->lru_head->lru_prev = entry;
    } else {
        cache->lru_tail = entry;
    }
    cache->lru_head = entry;
}

static void lru_move_to_head(page_cache_t *cache, cache_entry_t *entry) {
    if (entry == cache->lru_head) {
        return;
    }
    lru_remove(cache, entry);
    lru_add_to_head(cache, entry);
}

static cache_entry_t *cache_evict(page_cache_t *cache) {
    cache_entry_t *entry;
    cache_entry_t *prev;
    uint32_t hash_idx;

    entry = cache->lru_tail;
    while (entry) {
        if (entry->ref_count == 0) {
            lru_remove(cache, entry);
            hash_idx = hash_func(entry->page_num);
            prev = NULL;
            cache_entry_t *cur = cache->hash_table[hash_idx];
            while (cur) {
                if (cur->page_num == entry->page_num) {
                    if (prev) {
                        prev->next = cur->next;
                    } else {
                        cache->hash_table[hash_idx] = cur->next;
                    }
                    if (cur->next) {
                        cur->next->prev = prev;
                    }
                    break;
                }
                prev = cur;
                cur = cur->next;
            }
            cache->size--;
            return entry;
        }
        entry = entry->lru_prev;
    }
    return NULL;
}

void cache_init(page_cache_t *cache, int capacity) {
    int i;

    memset(cache, 0, sizeof(page_cache_t));
    for (i = 0; i < HASH_TABLE_SIZE; i++) {
        cache->hash_table[i] = NULL;
    }
    cache->size = 0;
    cache->capacity = capacity;
    cache->lru_head = NULL;
    cache->lru_tail = NULL;
    pthread_mutex_init(&cache->mutex, NULL);
}

void cache_destroy(page_cache_t *cache) {
    cache_entry_t *entry, *next;
    int i;

    pthread_mutex_lock(&cache->mutex);
    for (i = 0; i < HASH_TABLE_SIZE; i++) {
        entry = cache->hash_table[i];
        while (entry) {
            next = entry->next;
            cache_free_entry(entry);
            entry = next;
        }
        cache->hash_table[i] = NULL;
    }
    cache->size = 0;
    cache->lru_head = NULL;
    cache->lru_tail = NULL;
    pthread_mutex_unlock(&cache->mutex);
    pthread_mutex_destroy(&cache->mutex);
}

cache_entry_t *cache_new_entry(page_num_t page_num) {
    cache_entry_t *entry;

    entry = (cache_entry_t *)calloc(1, sizeof(cache_entry_t));
    if (!entry) {
        return NULL;
    }
    entry->page_num = page_num;
    entry->is_dirty = 0;
    entry->ref_count = 0;
    entry->last_access = time(NULL);
    entry->prev = NULL;
    entry->next = NULL;
    entry->lru_prev = NULL;
    entry->lru_next = NULL;

    return entry;
}

void cache_free_entry(cache_entry_t *entry) {
    if (entry) {
        free(entry);
    }
}

cache_entry_t *cache_get(btree_t *btree, page_num_t page_num) {
    page_cache_t *cache;
    uint32_t hash_idx;
    cache_entry_t *entry;

    cache = &btree->cache;
    pthread_mutex_lock(&cache->mutex);

    hash_idx = hash_func(page_num);
    entry = cache->hash_table[hash_idx];
    while (entry) {
        if (entry->page_num == page_num) {
            entry->ref_count++;
            entry->last_access = time(NULL);
            lru_move_to_head(cache, entry);
            pthread_mutex_unlock(&cache->mutex);
            return entry;
        }
        entry = entry->next;
    }

    entry = cache_new_entry(page_num);
    if (!entry) {
        pthread_mutex_unlock(&cache->mutex);
        return NULL;
    }

    if (page_read(btree->data_fd, page_num, &entry->page) < 0) {
        cache_free_entry(entry);
        pthread_mutex_unlock(&cache->mutex);
        return NULL;
    }

    if (cache->size >= cache->capacity) {
        cache_entry_t *evicted = cache_evict(cache);
        if (evicted) {
            if (evicted->is_dirty) {
                page_write(btree->data_fd, evicted->page_num, &evicted->page);
            }
            cache_free_entry(evicted);
        }
    }

    entry->ref_count = 1;
    entry->is_dirty = 0;
    entry->last_access = time(NULL);

    entry->next = cache->hash_table[hash_idx];
    if (cache->hash_table[hash_idx]) {
        cache->hash_table[hash_idx]->prev = entry;
    }
    entry->prev = NULL;
    cache->hash_table[hash_idx] = entry;
    cache->size++;

    lru_add_to_head(cache, entry);

    pthread_mutex_unlock(&cache->mutex);
    return entry;
}

void cache_put(btree_t *btree, cache_entry_t *entry) {
    page_cache_t *cache;
    uint32_t hash_idx;
    cache_entry_t *existing;

    cache = &btree->cache;
    pthread_mutex_lock(&cache->mutex);

    hash_idx = hash_func(entry->page_num);
    existing = cache->hash_table[hash_idx];
    while (existing) {
        if (existing->page_num == entry->page_num) {
            pthread_mutex_unlock(&cache->mutex);
            return;
        }
        existing = existing->next;
    }

    if (cache->size >= cache->capacity) {
        cache_entry_t *evicted = cache_evict(cache);
        if (evicted) {
            if (evicted->is_dirty) {
                page_write(btree->data_fd, evicted->page_num, &evicted->page);
            }
            cache_free_entry(evicted);
        }
    }

    entry->ref_count = 1;
    entry->last_access = time(NULL);

    entry->next = cache->hash_table[hash_idx];
    if (cache->hash_table[hash_idx]) {
        cache->hash_table[hash_idx]->prev = entry;
    }
    entry->prev = NULL;
    cache->hash_table[hash_idx] = entry;
    cache->size++;

    lru_add_to_head(cache, entry);

    pthread_mutex_unlock(&cache->mutex);
}

void cache_release(btree_t *btree, cache_entry_t *entry) {
    page_cache_t *cache;

    cache = &btree->cache;
    pthread_mutex_lock(&cache->mutex);

    if (entry->ref_count > 0) {
        entry->ref_count--;
    }

    pthread_mutex_unlock(&cache->mutex);
}

void cache_mark_dirty(btree_t *btree, cache_entry_t *entry) {
    page_cache_t *cache;

    cache = &btree->cache;
    pthread_mutex_lock(&cache->mutex);
    entry->is_dirty = 1;
    pthread_mutex_unlock(&cache->mutex);
}

int cache_flush(btree_t *btree) {
    page_cache_t *cache;
    cache_entry_t *entry;
    int i;

    cache = &btree->cache;
    pthread_mutex_lock(&cache->mutex);

    for (i = 0; i < HASH_TABLE_SIZE; i++) {
        entry = cache->hash_table[i];
        while (entry) {
            if (entry->is_dirty) {
                if (page_write(btree->data_fd, entry->page_num, &entry->page) < 0) {
                    pthread_mutex_unlock(&cache->mutex);
                    return -1;
                }
                entry->is_dirty = 0;
            }
            entry = entry->next;
        }
    }

    if (fsync(btree->data_fd) < 0) {
        pthread_mutex_unlock(&cache->mutex);
        return -1;
    }

    pthread_mutex_unlock(&cache->mutex);
    return 0;
}

int wal_open(wal_t *wal, const char *path, int create) {
    int flags;

    flags = O_RDWR;
    if (create) {
        flags |= O_CREAT;
    }

    wal->fd = open(path, flags, 0644);
    if (wal->fd < 0) {
        return -1;
    }

    wal->next_lsn = 1;
    wal->last_lsn = 0;
    pthread_mutex_init(&wal->mutex, NULL);

    return 0;
}

int wal_close(wal_t *wal) {
    if (wal->fd >= 0) {
        fsync(wal->fd);
        close(wal->fd);
        wal->fd = -1;
    }
    pthread_mutex_destroy(&wal->mutex);
    return 0;
}

int wal_sync(wal_t *wal) {
    pthread_mutex_lock(&wal->mutex);
    if (fsync(wal->fd) < 0) {
        pthread_mutex_unlock(&wal->mutex);
        return -1;
    }
    pthread_mutex_unlock(&wal->mutex);
    return 0;
}

uint32_t wal_append(wal_t *wal, wal_op_type_t op_type, page_num_t page_num,
                    const bt_key_t *key, const bt_value_t *value, uint16_t value_offset) {
    wal_header_t header;
    ssize_t nwritten;
    uint32_t lsn;

    pthread_mutex_lock(&wal->mutex);

    lsn = wal->next_lsn;
    header.lsn = lsn;
    header.prev_lsn = wal->last_lsn;
    header.op_type = op_type;
    header.page_num = page_num;
    header.key_len = key ? key->key_len : 0;
    header.value_len = value ? value->value_len : 0;
    header.value_offset = value_offset;

    nwritten = write(wal->fd, &header, sizeof(wal_header_t));
    if (nwritten != sizeof(wal_header_t)) {
        pthread_mutex_unlock(&wal->mutex);
        return 0;
    }

    if (key && key->key_len > 0) {
        nwritten = write(wal->fd, key->key_data, key->key_len);
        if (nwritten != (ssize_t)key->key_len) {
            pthread_mutex_unlock(&wal->mutex);
            return 0;
        }
    }

    if (value && value->value_len > 0) {
        nwritten = write(wal->fd, value->value_data, value->value_len);
        if (nwritten != (ssize_t)value->value_len) {
            pthread_mutex_unlock(&wal->mutex);
            return 0;
        }
    }

    wal->last_lsn = lsn;
    wal->next_lsn++;

    pthread_mutex_unlock(&wal->mutex);
    return lsn;
}

int wal_read(wal_t *wal, uint32_t lsn, wal_header_t *out_header,
             bt_key_t *out_key, bt_value_t *out_value) {
    off_t offset;
    ssize_t nread;

    (void)wal;
    (void)lsn;
    (void)out_header;
    (void)out_key;
    (void)out_value;

    offset = 0;
    if (lseek(wal->fd, offset, SEEK_SET) < 0) {
        return -1;
    }

    while (1) {
        nread = read(wal->fd, out_header, sizeof(wal_header_t));
        if (nread != sizeof(wal_header_t)) {
            break;
        }

        if (out_header->lsn == lsn) {
            if (out_key && out_header->key_len > 0) {
                out_key->key_len = out_header->key_len;
                nread = read(wal->fd, out_key->key_data, out_header->key_len);
                if (nread != (ssize_t)out_header->key_len) {
                    return -1;
                }
            } else if (out_header->key_len > 0) {
                if (lseek(wal->fd, out_header->key_len, SEEK_CUR) < 0) {
                    return -1;
                }
            }

            if (out_value && out_header->value_len > 0) {
                out_value->value_len = out_header->value_len;
                nread = read(wal->fd, out_value->value_data, out_header->value_len);
                if (nread != (ssize_t)out_header->value_len) {
                    return -1;
                }
            } else if (out_header->value_len > 0) {
                if (lseek(wal->fd, out_header->value_len, SEEK_CUR) < 0) {
                    return -1;
                }
            }

            return 0;
        } else {
            if (out_header->key_len > 0) {
                if (lseek(wal->fd, out_header->key_len, SEEK_CUR) < 0) {
                    return -1;
                }
            }
            if (out_header->value_len > 0) {
                if (lseek(wal->fd, out_header->value_len, SEEK_CUR) < 0) {
                    return -1;
                }
            }
        }
    }

    return -1;
}
