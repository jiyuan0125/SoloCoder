#ifndef BTREE_CACHE_H
#define BTREE_CACHE_H

#include "common.h"
#include "page.h"

void cache_init(page_cache_t *cache, int capacity);
void cache_destroy(page_cache_t *cache);

cache_entry_t *cache_get(btree_t *btree, page_num_t page_num);
void cache_put(btree_t *btree, cache_entry_t *entry);
void cache_release(btree_t *btree, cache_entry_t *entry);

void cache_mark_dirty(btree_t *btree, cache_entry_t *entry);
int cache_flush(btree_t *btree);

cache_entry_t *cache_new_entry(page_num_t page_num);
void cache_free_entry(cache_entry_t *entry);

int wal_open(wal_t *wal, const char *path, int create);
int wal_close(wal_t *wal);

uint32_t wal_append(wal_t *wal, wal_op_type_t op_type, page_num_t page_num,
                    const bt_key_t *key, const bt_value_t *value, uint16_t value_offset);
int wal_read(wal_t *wal, uint32_t lsn, wal_header_t *out_header,
             bt_key_t *out_key, bt_value_t *out_value);

int wal_sync(wal_t *wal);

#endif
