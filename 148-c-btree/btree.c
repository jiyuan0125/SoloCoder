#include "btree.h"

typedef struct {
    page_num_t page_num;
} path_item_t;

static int btree_insert_internal(btree_t *btree, page_num_t page_num,
                                   const bt_key_t *key, page_num_t left_child,
                                   page_num_t right_child, path_item_t *path,
                                   int path_len);

static int btree_delete_internal(btree_t *btree, page_num_t page_num,
                                   const bt_key_t *key, path_item_t *path,
                                   int path_len);

int btree_create(btree_t *btree, const char *data_path, const char *wal_path) {
    int fd;
    page_t root_page;

    memset(btree, 0, sizeof(btree_t));

    fd = open(data_path, O_RDWR | O_CREAT | O_TRUNC, 0644);
    if (fd < 0) {
        return -1;
    }
    btree->data_fd = fd;

    memset(&btree->header, 0, sizeof(file_header_t));
    btree->header.magic = BTREE_MAGIC;
    btree->header.version = 1;
    btree->header.root_page = ROOT_PAGE_INITIAL;
    btree->header.num_pages = 2;
    btree->header.freelist_head = INVALID_PAGE_NUM;

    if (btree_write_header(btree) < 0) {
        close(fd);
        return -1;
    }

    page_init_leaf(&root_page, ROOT_PAGE_INITIAL);
    if (page_write(fd, ROOT_PAGE_INITIAL, &root_page) < 0) {
        close(fd);
        return -1;
    }

    cache_init(&btree->cache, CACHE_CAPACITY);

    if (wal_open(&btree->wal, wal_path, 1) < 0) {
        cache_destroy(&btree->cache);
        close(fd);
        return -1;
    }

    pthread_rwlock_init(&btree->tree_lock, NULL);
    btree->is_open = 1;

    return 0;
}

int btree_open(btree_t *btree, const char *data_path, const char *wal_path) {
    int fd;

    memset(btree, 0, sizeof(btree_t));

    fd = open(data_path, O_RDWR, 0644);
    if (fd < 0) {
        return -1;
    }
    btree->data_fd = fd;

    if (btree_read_header(btree) < 0) {
        close(fd);
        return -1;
    }

    if (btree->header.magic != BTREE_MAGIC) {
        close(fd);
        return -1;
    }

    cache_init(&btree->cache, CACHE_CAPACITY);
    pthread_rwlock_init(&btree->tree_lock, NULL);
    btree->is_open = 1;

    wal_recover(btree, wal_path);

    if (wal_open(&btree->wal, wal_path, 0) < 0) {
        cache_destroy(&btree->cache);
        pthread_rwlock_destroy(&btree->tree_lock);
        close(fd);
        btree->is_open = 0;
        return -1;
    }

    return 0;
}

int btree_close(btree_t *btree) {
    if (!btree->is_open) {
        return 0;
    }

    cache_flush(btree);
    wal_sync(&btree->wal);

    wal_close(&btree->wal);
    cache_destroy(&btree->cache);

    if (btree->data_fd >= 0) {
        fsync(btree->data_fd);
        close(btree->data_fd);
        btree->data_fd = -1;
    }

    pthread_rwlock_destroy(&btree->tree_lock);
    btree->is_open = 0;

    return 0;
}

int btree_get(btree_t *btree, const bt_key_t *key, bt_value_t *out_value) {
    page_num_t current_page_num;
    cache_entry_t *entry;
    page_t *page;
    int found;
    uint16_t child_idx;
    leaf_entry_t *leaf_entry;
    int result;

    if (!key || key->key_len == 0 || key->key_len > MAX_KEY_SIZE) {
        return -1;
    }

    pthread_rwlock_rdlock(&btree->tree_lock);

    current_page_num = btree->header.root_page;
    result = -1;

    while (1) {
        entry = cache_get(btree, current_page_num);
        if (!entry) {
            pthread_rwlock_unlock(&btree->tree_lock);
            return -1;
        }
        page = &entry->page;

        if (page->type == PAGE_TYPE_LEAF) {
            uint16_t idx;
            found = page_find_leaf(page, key, &idx);
            if (found) {
                leaf_entry = page_get_leaf_entry(page, idx);
                if (leaf_entry) {
                    if (out_value) {
                        if (value_page_read(btree->data_fd, leaf_entry->value_page, out_value) < 0) {
                            cache_release(btree, entry);
                            pthread_rwlock_unlock(&btree->tree_lock);
                            return -1;
                        }
                    }
                    result = 0;
                } else if (found) {
                    result = 0;
                }
            } else {
                result = -1;
            }
            cache_release(btree, entry);
            break;
        } else {
            page_find_internal(page, key, &child_idx);
            current_page_num = page_get_child(page, child_idx);
            if (current_page_num == INVALID_PAGE_NUM) {
                cache_release(btree, entry);
                result = -1;
                break;
            }
            cache_release(btree, entry);
        }
    }

    pthread_rwlock_unlock(&btree->tree_lock);
    return result;
}

static int btree_insert_internal(btree_t *btree, page_num_t page_num,
                                   const bt_key_t *key, page_num_t left_child,
                                   page_num_t right_child, path_item_t *path,
                                   int path_len) {
    cache_entry_t *entry;
    page_t *page;
    uint16_t max_entries;
    int ret;
    bt_key_t mid_key;
    page_num_t new_page_num;
    cache_entry_t *new_entry;
    path_item_t parent_item;
    int parent_path_idx;

    entry = cache_get(btree, page_num);
    if (!entry) {
        return -1;
    }
    page = &entry->page;

    max_entries = page_get_max_entries_internal();

    if (page->num_entries < max_entries) {
        ret = page_insert_internal(page, key, left_child, right_child);
        if (ret == 0) {
            cache_mark_dirty(btree, entry);
        }
        cache_release(btree, entry);
        return ret;
    }

    new_page_num = page_allocate(btree);
    if (new_page_num == INVALID_PAGE_NUM) {
        cache_release(btree, entry);
        return -1;
    }

    new_entry = cache_new_entry(new_page_num);
    if (!new_entry) {
        cache_release(btree, entry);
        return -1;
    }
    page_init_internal(&new_entry->page, new_page_num);

    page_insert_internal(page, key, left_child, right_child);

    ret = page_split_internal(page, &new_entry->page, &mid_key);
    if (ret < 0) {
        cache_free_entry(new_entry);
        cache_release(btree, entry);
        return -1;
    }

    cache_mark_dirty(btree, entry);
    cache_put(btree, new_entry);
    cache_mark_dirty(btree, new_entry);

    if (path_len > 0) {
        parent_path_idx = path_len - 1;
        parent_item = path[parent_path_idx];

        ret = btree_insert_internal(btree, parent_item.page_num, &mid_key,
                                     page_num, new_page_num,
                                     path, parent_path_idx);
    } else {
        page_num_t new_root_num;
        cache_entry_t *new_root_entry;

        new_root_num = page_allocate(btree);
        if (new_root_num == INVALID_PAGE_NUM) {
            cache_release(btree, entry);
            cache_release(btree, new_entry);
            return -1;
        }

        new_root_entry = cache_new_entry(new_root_num);
        if (!new_root_entry) {
            cache_release(btree, entry);
            cache_release(btree, new_entry);
            return -1;
        }
        page_init_internal(&new_root_entry->page, new_root_num);

        page_insert_internal(&new_root_entry->page, &mid_key, page_num, new_page_num);

        btree->header.root_page = new_root_num;
        btree_write_header(btree);

        cache_mark_dirty(btree, new_root_entry);
        cache_put(btree, new_root_entry);
        cache_release(btree, new_root_entry);

        ret = 0;
    }

    cache_release(btree, entry);
    cache_release(btree, new_entry);

    return ret;
}

int btree_put(btree_t *btree, const bt_key_t *key, const bt_value_t *value) {
    page_num_t current_page_num;
    cache_entry_t *entry;
    page_t *page;
    uint16_t child_idx;
    int found;
    uint16_t max_entries;
    path_item_t path[64];
    int path_len = 0;
    int ret;
    bt_key_t mid_key;
    page_num_t new_page_num;
    cache_entry_t *new_entry;
    page_num_t value_page;

    if (!key || key->key_len == 0 || key->key_len > MAX_KEY_SIZE) {
        return -1;
    }
    if (!value || value->value_len > MAX_VALUE_SIZE) {
        return -1;
    }

    pthread_rwlock_wrlock(&btree->tree_lock);

    current_page_num = btree->header.root_page;
    path_len = 0;

    while (1) {
        entry = cache_get(btree, current_page_num);
        if (!entry) {
            pthread_rwlock_unlock(&btree->tree_lock);
            return -1;
        }
        page = &entry->page;

        path[path_len].page_num = current_page_num;
        path_len++;

        if (page->type == PAGE_TYPE_LEAF) {
            uint16_t idx;
            found = page_find_leaf(page, key, &idx);
            max_entries = page_get_max_entries_leaf();

            if (found) {
                leaf_entry_t *leaf_entry = page_get_leaf_entry(page, idx);
                if (leaf_entry) {
                    if (value_page_write(btree->data_fd, leaf_entry->value_page, value) < 0) {
                        cache_release(btree, entry);
                        pthread_rwlock_unlock(&btree->tree_lock);
                        return -1;
                    }
                }
                cache_release(btree, entry);
                ret = 0;
                break;
            }

            value_page = page_allocate(btree);
            if (value_page == INVALID_PAGE_NUM) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            if (value_page_write(btree->data_fd, value_page, value) < 0) {
                page_deallocate(btree, value_page);
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            if (page->num_entries < max_entries) {
                page_insert_leaf(page, key, value_page, value->value_len);
                cache_mark_dirty(btree, entry);
                cache_release(btree, entry);
                ret = 0;
                break;
            }

            new_page_num = page_allocate(btree);
            if (new_page_num == INVALID_PAGE_NUM) {
                page_deallocate(btree, value_page);
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            new_entry = cache_new_entry(new_page_num);
            if (!new_entry) {
                page_deallocate(btree, value_page);
                page_deallocate(btree, new_page_num);
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }
            page_init_leaf(&new_entry->page, new_page_num);

            page_split_leaf(page, &new_entry->page, &mid_key);

            if (key_compare(key, &mid_key) < 0) {
                page_insert_leaf(page, key, value_page, value->value_len);
            } else {
                page_insert_leaf(&new_entry->page, key, value_page, value->value_len);
            }

            cache_mark_dirty(btree, entry);
            cache_put(btree, new_entry);
            cache_mark_dirty(btree, new_entry);

            if (path_len > 1) {
                path_item_t parent_item = path[path_len - 2];
                ret = btree_insert_internal(btree, parent_item.page_num,
                                             &mid_key, current_page_num,
                                             new_page_num, path, path_len - 2);
            } else {
                page_num_t new_root_num;
                cache_entry_t *new_root_entry;

                new_root_num = page_allocate(btree);
                if (new_root_num == INVALID_PAGE_NUM) {
                    page_deallocate(btree, value_page);
                    cache_release(btree, entry);
                    cache_release(btree, new_entry);
                    pthread_rwlock_unlock(&btree->tree_lock);
                    return -1;
                }

                new_root_entry = cache_new_entry(new_root_num);
                if (!new_root_entry) {
                    page_deallocate(btree, value_page);
                    page_deallocate(btree, new_root_num);
                    cache_release(btree, entry);
                    cache_release(btree, new_entry);
                    pthread_rwlock_unlock(&btree->tree_lock);
                    return -1;
                }
                page_init_internal(&new_root_entry->page, new_root_num);

                page_insert_internal(&new_root_entry->page, &mid_key, current_page_num, new_page_num);

                btree->header.root_page = new_root_num;
                btree_write_header(btree);

                cache_mark_dirty(btree, new_root_entry);
                cache_put(btree, new_root_entry);
                cache_release(btree, new_root_entry);
                ret = 0;
            }

            cache_release(btree, entry);
            cache_release(btree, new_entry);
            break;
        } else {
            page_find_internal(page, key, &child_idx);
            current_page_num = page_get_child(page, child_idx);
            if (current_page_num == INVALID_PAGE_NUM) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }
            cache_release(btree, entry);
        }
    }

    if (ret == 0) {
        wal_append(&btree->wal, WAL_OP_TYPE_INSERT, INVALID_PAGE_NUM,
                    key, value, 0);
        wal_sync(&btree->wal);
    }

    pthread_rwlock_unlock(&btree->tree_lock);
    return ret;
}

static int btree_delete_recursive(btree_t *btree, page_num_t page_num,
                                    const bt_key_t *key, path_item_t *path,
                                    int path_len, int *deleted_key_is_first);

int btree_delete(btree_t *btree, const bt_key_t *key) {
    path_item_t path[64];
    int path_len = 0;
    int deleted_key_is_first = 0;
    int ret;

    if (!key || key->key_len == 0 || key->key_len > MAX_KEY_SIZE) {
        return -1;
    }

    pthread_rwlock_wrlock(&btree->tree_lock);

    ret = btree_delete_recursive(btree, btree->header.root_page, key,
                                  path, path_len, &deleted_key_is_first);

    if (ret == 0) {
        wal_append(&btree->wal, WAL_OP_TYPE_DELETE, INVALID_PAGE_NUM,
                    key, NULL, 0);
        wal_sync(&btree->wal);
    }

    pthread_rwlock_unlock(&btree->tree_lock);
    return ret;
}

static int btree_delete_recursive(btree_t *btree, page_num_t page_num,
                                    const bt_key_t *key, path_item_t *path,
                                    int path_len, int *deleted_key_is_first) {
    cache_entry_t *entry;
    page_t *page;
    uint16_t idx;
    int found;
    int ret = -1;

    entry = cache_get(btree, page_num);
    if (!entry) {
        return -1;
    }
    page = &entry->page;

    path[path_len].page_num = page_num;
    path_len++;

    if (page->type == PAGE_TYPE_LEAF) {
        found = page_find_leaf(page, key, &idx);
        if (!found) {
            cache_release(btree, entry);
            return -1;
        }

        leaf_entry_t *leaf_entry = page_get_leaf_entry(page, idx);
        if (leaf_entry) {
            page_deallocate(btree, leaf_entry->value_page);
        }

        *deleted_key_is_first = (idx == 0);

        page_remove_leaf(page, idx);
        cache_mark_dirty(btree, entry);

        if (page->num_entries == 0 && page_num != btree->header.root_page) {
            page_num_t prev_leaf = page->prev_leaf;
            page_num_t next_leaf = page->next_leaf;

            if (prev_leaf != INVALID_PAGE_NUM) {
                cache_entry_t *prev_entry = cache_get(btree, prev_leaf);
                if (prev_entry) {
                    prev_entry->page.next_leaf = next_leaf;
                    cache_mark_dirty(btree, prev_entry);
                    cache_release(btree, prev_entry);
                }
            }
            if (next_leaf != INVALID_PAGE_NUM) {
                cache_entry_t *next_entry = cache_get(btree, next_leaf);
                if (next_entry) {
                    next_entry->page.prev_leaf = prev_leaf;
                    cache_mark_dirty(btree, next_entry);
                    cache_release(btree, next_entry);
                }
            }

            cache_release(btree, entry);
            page_deallocate(btree, page_num);
            ret = 1;
        } else {
            cache_release(btree, entry);
            ret = 0;
        }

        return ret;
    } else {
        page_find_internal(page, key, &idx);
        page_num_t child_num = page_get_child(page, idx);

        if (child_num == INVALID_PAGE_NUM) {
            cache_release(btree, entry);
            return -1;
        }

        int child_key_is_first = 0;
        ret = btree_delete_recursive(btree, child_num, key, path, path_len,
                                      &child_key_is_first);

        if (ret < 0) {
            cache_release(btree, entry);
            return ret;
        }

        if (child_key_is_first && idx == 0 && page->num_entries > 0) {
            cache_entry_t *child_entry = cache_get(btree, child_num);
            if (child_entry) {
                bt_key_t new_min_key;
                if (page_get_min_key(&child_entry->page, &new_min_key) == 0) {
                    internal_entry_t *first_entry = (internal_entry_t *)(page->data + sizeof(page_num_t));
                    key_copy(&first_entry->key, &new_min_key);
                    cache_mark_dirty(btree, entry);
                }
                cache_release(btree, child_entry);
            }
        }

        if (ret == 1) {
            uint16_t child_idx_in_parent;
            if (page_find_child_idx(page, child_num, &child_idx_in_parent) == 0) {
                if (child_idx_in_parent == 0) {
                    page_num_t next_child = page_get_child(page, 1);
                    page_set_child(page, 0, next_child);
                    page_remove_internal(page, 0);
                } else {
                    page_remove_internal(page, child_idx_in_parent - 1);
                }
                cache_mark_dirty(btree, entry);
            }

            if (page->num_entries == 0) {
                if (page_num == btree->header.root_page) {
                    page_num_t first_child = page_get_child(page, 0);
                    if (first_child != INVALID_PAGE_NUM) {
                        btree->header.root_page = first_child;
                        btree_write_header(btree);
                    }
                    cache_release(btree, entry);
                    page_deallocate(btree, page_num);
                    return 0;
                } else {
                    cache_release(btree, entry);
                    page_deallocate(btree, page_num);
                    return 1;
                }
            }
        }

        cache_release(btree, entry);
        return 0;
    }
}

static int btree_delete_internal(btree_t *btree, page_num_t page_num,
                                   const bt_key_t *key, path_item_t *path,
                                   int path_len) {
    (void)btree;
    (void)page_num;
    (void)key;
    (void)path;
    (void)path_len;
    return 0;
}

int btree_update_partial(btree_t *btree, const bt_key_t *key,
                          uint16_t offset, const uint8_t *data, uint16_t len) {
    page_num_t current_page_num;
    cache_entry_t *entry;
    page_t *page;
    uint16_t child_idx;
    uint16_t idx;
    int found;
    int ret;
    leaf_entry_t *leaf_entry;
    bt_value_t wal_value;
    bt_value_t current_value;

    if (!key || key->key_len == 0 || key->key_len > MAX_KEY_SIZE) {
        return -1;
    }
    if (!data || len == 0 || offset + len > MAX_VALUE_SIZE) {
        return -1;
    }

    pthread_rwlock_wrlock(&btree->tree_lock);

    current_page_num = btree->header.root_page;

    while (1) {
        entry = cache_get(btree, current_page_num);
        if (!entry) {
            pthread_rwlock_unlock(&btree->tree_lock);
            return -1;
        }
        page = &entry->page;

        if (page->type == PAGE_TYPE_LEAF) {
            found = page_find_leaf(page, key, &idx);
            if (!found) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            leaf_entry = page_get_leaf_entry(page, idx);
            if (!leaf_entry) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            if (value_page_read(btree->data_fd, leaf_entry->value_page, &current_value) < 0) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            if (offset + len > current_value.value_len) {
                if (offset + len > MAX_VALUE_SIZE) {
                    cache_release(btree, entry);
                    pthread_rwlock_unlock(&btree->tree_lock);
                    return -1;
                }
                current_value.value_len = offset + len;
            }

            memcpy(&current_value.value_data[offset], data, len);

            if (value_page_write(btree->data_fd, leaf_entry->value_page, &current_value) < 0) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }

            cache_release(btree, entry);
            ret = 0;
            break;
        } else {
            page_find_internal(page, key, &child_idx);
            current_page_num = page_get_child(page, child_idx);
            if (current_page_num == INVALID_PAGE_NUM) {
                cache_release(btree, entry);
                pthread_rwlock_unlock(&btree->tree_lock);
                return -1;
            }
            cache_release(btree, entry);
        }
    }

    if (ret == 0) {
        wal_append(&btree->wal, WAL_OP_TYPE_UPDATE, INVALID_PAGE_NUM,
                    key, &current_value, 0);
        wal_sync(&btree->wal);
    }

    pthread_rwlock_unlock(&btree->tree_lock);
    return ret;
}

int btree_flush(btree_t *btree) {
    return cache_flush(btree);
}

int btree_sync(btree_t *btree) {
    int ret;

    ret = cache_flush(btree);
    if (ret < 0) {
        return ret;
    }

    ret = wal_sync(&btree->wal);
    return ret;
}

int wal_recover(btree_t *btree, const char *wal_path) {
    int wal_fd;
    wal_header_t header;
    ssize_t nread;
    bt_key_t key;
    bt_value_t value;
    off_t file_size;

    wal_fd = open(wal_path, O_RDWR, 0644);
    if (wal_fd < 0) {
        return 0;
    }

    file_size = lseek(wal_fd, 0, SEEK_END);
    if (file_size <= 0) {
        close(wal_fd);
        return 0;
    }
    lseek(wal_fd, 0, SEEK_SET);

    while (1) {
        nread = read(wal_fd, &header, sizeof(wal_header_t));
        if (nread != sizeof(wal_header_t)) {
            break;
        }

        if (header.key_len > 0) {
            key.key_len = header.key_len;
            nread = read(wal_fd, key.key_data, header.key_len);
            if (nread != (ssize_t)header.key_len) {
                close(wal_fd);
                return -1;
            }
        }

        if (header.value_len > 0) {
            value.value_len = header.value_len;
            nread = read(wal_fd, value.value_data, header.value_len);
            if (nread != (ssize_t)header.value_len) {
                close(wal_fd);
                return -1;
            }
        }

        if (header.op_type == WAL_OP_TYPE_INSERT || header.op_type == WAL_OP_TYPE_UPDATE) {
            if (header.key_len > 0 && header.value_len > 0) {
                btree_put(btree, &key, &value);
            }
        } else if (header.op_type == WAL_OP_TYPE_DELETE) {
            if (header.key_len > 0) {
                btree_delete(btree, &key);
            }
        }
    }

    cache_flush(btree);

    if (ftruncate(wal_fd, 0) < 0) {
        close(wal_fd);
        return -1;
    }
    fsync(wal_fd);
    close(wal_fd);

    return 0;
}
