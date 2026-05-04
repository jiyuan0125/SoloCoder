#ifndef BTREE_BTREE_H
#define BTREE_BTREE_H

#include "common.h"
#include "page.h"
#include "cache.h"

#define BTREE_MAGIC 0x42545245

int btree_open(btree_t *btree, const char *data_path, const char *wal_path);
int btree_close(btree_t *btree);
int btree_create(btree_t *btree, const char *data_path, const char *wal_path);

int btree_get(btree_t *btree, const bt_key_t *key, bt_value_t *out_value);
int btree_put(btree_t *btree, const bt_key_t *key, const bt_value_t *value);
int btree_delete(btree_t *btree, const bt_key_t *key);

int btree_update_partial(btree_t *btree, const bt_key_t *key,
                          uint16_t offset, const uint8_t *data, uint16_t len);

int btree_flush(btree_t *btree);
int btree_sync(btree_t *btree);

int wal_recover(btree_t *btree, const char *wal_path);

#endif
