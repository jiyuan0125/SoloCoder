#ifndef BTREE_PAGE_H
#define BTREE_PAGE_H

#include "common.h"

void page_init_internal(page_t *page, page_num_t page_num);
void page_init_leaf(page_t *page, page_num_t page_num);

static inline page_num_t *page_get_first_child_ptr(page_t *page) {
    if (page->type != PAGE_TYPE_INTERNAL) {
        return NULL;
    }
    return (page_num_t *)page->data;
}

static inline internal_entry_t *page_get_internal_entry_ptr(page_t *page, uint16_t idx) {
    if (page->type != PAGE_TYPE_INTERNAL || idx >= page->num_entries) {
        return NULL;
    }
    return (internal_entry_t *)(page->data + sizeof(page_num_t)) + idx;
}

static inline uint16_t page_get_max_entries_internal(void) {
    return (PAGE_SIZE - 20 - sizeof(page_num_t)) / sizeof(internal_entry_t);
}

static inline uint16_t page_get_max_entries_leaf(void) {
    return (PAGE_SIZE - 20) / sizeof(leaf_entry_t);
}

static inline page_num_t page_get_child(const page_t *page, uint16_t idx) {
    page_num_t first_child;
    const internal_entry_t *entry;

    if (page->type != PAGE_TYPE_INTERNAL) {
        return INVALID_PAGE_NUM;
    }

    if (idx == 0) {
        memcpy(&first_child, page->data, sizeof(page_num_t));
        return first_child;
    }

    if (idx > page->num_entries) {
        return INVALID_PAGE_NUM;
    }

    entry = (const internal_entry_t *)(page->data + sizeof(page_num_t)) + (idx - 1);
    return entry->child;
}

static inline int page_set_child(page_t *page, uint16_t idx, page_num_t child) {
    internal_entry_t *entry;

    if (page->type != PAGE_TYPE_INTERNAL) {
        return -1;
    }

    if (idx == 0) {
        memcpy(page->data, &child, sizeof(page_num_t));
        return 0;
    }

    if (idx > page->num_entries + 1) {
        return -1;
    }

    entry = (internal_entry_t *)(page->data + sizeof(page_num_t)) + (idx - 1);
    entry->child = child;
    return 0;
}

static inline leaf_entry_t *page_get_leaf_entry(page_t *page, uint16_t idx) {
    if (page->type != PAGE_TYPE_LEAF || idx >= page->num_entries) {
        return NULL;
    }
    return (leaf_entry_t *)page->data + idx;
}

int page_insert_internal(page_t *page, const bt_key_t *key, page_num_t left_child, page_num_t right_child);
int page_insert_leaf(page_t *page, const bt_key_t *key, page_num_t value_page, uint32_t value_len);
int page_update_leaf(page_t *page, uint16_t idx, page_num_t value_page, uint32_t value_len);

int page_remove_internal(page_t *page, uint16_t idx);
int page_remove_leaf(page_t *page, uint16_t idx);

int page_find_internal(const page_t *page, const bt_key_t *key, uint16_t *out_child_idx);
int page_find_leaf(const page_t *page, const bt_key_t *key, uint16_t *out_idx);

int page_split_internal(page_t *old_page, page_t *new_page, bt_key_t *out_mid_key);
int page_split_leaf(page_t *old_page, page_t *new_page, bt_key_t *out_mid_key);

int page_merge_internal(page_t *left_page, page_t *right_page, const bt_key_t *sep_key);
int page_merge_leaf(page_t *left_page, page_t *right_page);

int page_read(int fd, page_num_t page_num, page_t *page);
int page_write(int fd, page_num_t page_num, const page_t *page);

int value_page_read(int fd, page_num_t page_num, bt_value_t *value);
int value_page_write(int fd, page_num_t page_num, const bt_value_t *value);
int value_page_update_partial(int fd, page_num_t page_num, uint32_t offset, const uint8_t *data, uint32_t len);

page_num_t page_allocate(btree_t *btree);
void page_deallocate(btree_t *btree, page_num_t page_num);

int btree_read_header(btree_t *btree);
int btree_write_header(btree_t *btree);

#endif
